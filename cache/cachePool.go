package cache

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/panjf2000/goproxy/interface"
	"net/http"
	"time"
)

func MD5Uri(uri string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(uri)))
}

type CachePool struct {
	pool *pool.Pool
}

func NewCachePool(address, password string, cap int) *CachePool {
	p, err := pool.NewCustom("tcp", address, cap, func(network, addr string) (*redis.Client, error) {
		client, err := redis.Dial(network, addr)
		if err != nil {
			return nil, err
		}
		if err = client.Cmd("AUTH", password).Err; err != nil {
			client.Close()
			return nil, err
		}
		return client, nil
	})
	if err != nil {
		panic(err)
	}
	n := p.Avail()
	if n == 0 {

	}
	return &CachePool{pool: p}

}

func (c *CachePool) Get(uri string) api.Cache {
	//log.Println("get cahche of ", uri)
	if cache := c.get(MD5Uri(uri)); cache != nil {
		//log.Println(*cache)
		return cache
	}
	return nil
}

func (c *CachePool) get(md5Uri string) *Cache {
	conn, _ := c.pool.Get()
	defer c.pool.Put(conn)

	//b, err := redis.Bytes(conn.Do("GET", md5Uri))
	b, err := conn.Cmd("GET", md5Uri).Bytes()
	if err != nil || len(b) == 0 {
		//log.Println(err)
		return nil
	}
	//log.Println(string(b))
	cache := new(Cache)
	json.Unmarshal(b, &cache)
	return cache
}

func (c *CachePool) Delete(uri string) {
	c.delete(MD5Uri(uri))
}

func (c *CachePool) delete(md5Uri string) {
	conn, _ := c.pool.Get()
	defer c.pool.Put(conn)

	if err := conn.Cmd("DEL", md5Uri).Err; err != nil {
		return
	}
	return
}

func (c *CachePool) CheckAndStore(uri string, resp *http.Response) {
	if !IsCache(resp) {
		return
	}

	cache := New(resp)

	if cache == nil {
		return
	}

	//log.Println("store cache ", uri)

	md5Uri := MD5Uri(uri)
	b, err := json.Marshal(cache)
	if err != nil {
		//log.Println(err)
		return
	}

	conn, _ := c.pool.Get()
	defer c.pool.Put(conn)

	err = conn.Cmd("MULTI").Err
	//log.Println("successfully store cache ", uri)
	conn.Cmd("SET", md5Uri, b)
	conn.Cmd("EXPIRE", md5Uri, cache.maxAge)
	err = conn.Cmd("EXEC").Err
	if err != nil {
		//log.Println(err)
		return
	}

}

func (c *CachePool) Clear(d time.Duration) {

}
