package cache

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/panjf2000/goproxy/interface"
)

func MD5Uri(uri string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(uri)))
}

type ConnCachePool struct {
	pool *redis.Pool
}

func NewCachePool(address, password string, idleTimeout, cap, maxIdle int) *ConnCachePool {
	redisPool := &redis.Pool{
		MaxActive:   cap,
		MaxIdle:     maxIdle,
		IdleTimeout: time.Duration(idleTimeout) * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", address)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := conn.Do("AUTH", password); err != nil {
					conn.Close()
					return nil, err
				}
			}
			return conn, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				panic(err)
			}
			return err

		},
	}
	return &ConnCachePool{pool: redisPool}

}

func (c *ConnCachePool) Get(uri string) api.Cache {
	if respCache := c.get(MD5Uri(uri)); respCache != nil {
		return respCache
	}
	return nil
}

func (c *ConnCachePool) get(md5Uri string) *HttpCache {
	conn := c.pool.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", md5Uri))
	if err != nil || len(b) == 0 {
		return nil
	}
	respCache := new(HttpCache)
	json.Unmarshal(b, respCache)
	return respCache
}

func (c *ConnCachePool) Delete(uri string) {
	c.delete(MD5Uri(uri))
}

func (c *ConnCachePool) delete(md5Uri string) {
	conn := c.pool.Get()
	defer conn.Close()

	if _, err := conn.Do("DEL", md5Uri); err != nil {
		return
	}
	return
}

func (c *ConnCachePool) CheckAndStore(uri string, req *http.Request, resp *http.Response) {
	if !IsReqCache(req) || !IsRespCache(resp) {
		return
	}
	respCache := NewCacheResp(resp)

	if respCache == nil {
		return
	}

	md5Uri := MD5Uri(uri)
	b, err := json.Marshal(respCache)
	if err != nil {
		//log.Println(err)
		return
	}

	conn := c.pool.Get()
	defer conn.Close()

	_, err = conn.Do("MULTI")
	conn.Do("SET", md5Uri, b)
	conn.Do("EXPIRE", md5Uri, respCache.maxAge)
	_, err = conn.Do("EXEC")
	if err != nil {
		//log.Println(err)
		return
	}

}

func (c *ConnCachePool) Clear(d time.Duration) {

}
