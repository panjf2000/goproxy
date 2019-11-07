package cache

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/panjf2000/goproxy/interface"
	"github.com/panjf2000/goproxy/tool"
)

type RedisConnCachePool struct {
	pool *redis.Pool
}

func NewRedisCachePool(address, password string, idleTimeout, cap, maxIdle int) *RedisConnCachePool {
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
					_ = conn.Close()
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
	return &RedisConnCachePool{pool: redisPool}

}

func (c *RedisConnCachePool) Get(uri string) api.Cache {
	if respCache := c.get(tool.MD5Uri(uri)); respCache != nil {
		return respCache
	}
	return nil
}

func (c *RedisConnCachePool) get(md5Uri string) *HttpCache {
	conn := c.pool.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", md5Uri))
	if err != nil || len(b) == 0 {
		return nil
	}
	respCache := new(HttpCache)
	_ = json.Unmarshal(b, respCache)
	return respCache
}

func (c *RedisConnCachePool) Delete(uri string) {
	c.delete(tool.MD5Uri(uri))
}

func (c *RedisConnCachePool) delete(md5Uri string) {
	conn := c.pool.Get()
	defer conn.Close()

	if _, err := conn.Do("DEL", md5Uri); err != nil {
		return
	}
	return
}

func (c *RedisConnCachePool) CheckAndStore(uri string, req *http.Request, resp *http.Response) {
	if !IsReqCache(req) || !IsRespCache(resp) {
		return
	}
	respCache := NewCacheResp(resp)

	if respCache == nil {
		return
	}

	md5Uri := tool.MD5Uri(uri)
	b, err := json.Marshal(respCache)
	if err != nil {
		return
	}

	conn := c.pool.Get()
	defer conn.Close()

	_, err = conn.Do("MULTI")
	_, _ = conn.Do("SET", md5Uri, b)
	_, _ = conn.Do("EXPIRE", md5Uri, respCache.maxAge)
	_, err = conn.Do("EXEC")
	if err != nil {
		return
	}

}

//func (c *RedisConnCachePool) Clear(d time.Duration) {
//
//}
