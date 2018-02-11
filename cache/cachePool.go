package cache

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/panjf2000/goproxy/interface"
)

func MD5Uri(uri string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(uri)))
}

type ConnCachePool struct {
	pool *pool.Pool
}

//func HeartBeat(p *pool.Pool, intervalTime int) {
//	go func() {
//		for {
//			p.Cmd("PING")
//			time.Sleep(time.Duration(intervalTime) * time.Second)
//		}
//	}()
//}

func NewCachePool(address, password string, cap int) *ConnCachePool {
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

	// keep redis pool alive, according to the author of radix.v2,it's unnecessary to this anymore in the new version
	// cuz it will do it automatically.
	//HeartBeat(p, 10)

	return &ConnCachePool{pool: p}

}

func (c *ConnCachePool) Get(uri string) api.Cache {
	//log.Println("get cahche of ", uri)
	if respCache := c.get(MD5Uri(uri)); respCache != nil {
		//log.Println(*cacheResp)
		return respCache
	}
	return nil
}

func (c *ConnCachePool) get(md5Uri string) *HttpCache {
	conn, _ := c.pool.Get()
	defer c.pool.Put(conn)

	//b, err := redis.Bytes(conn.Do("GET", md5Uri))
	b, err := conn.Cmd("GET", md5Uri).Bytes()
	if err != nil || len(b) == 0 {
		//log.Println(err)
		return nil
	}
	//log.Println(string(b))
	respCache := new(HttpCache)
	json.Unmarshal(b, &respCache)
	return respCache
}

func (c *ConnCachePool) Delete(uri string) {
	c.delete(MD5Uri(uri))
}

func (c *ConnCachePool) delete(md5Uri string) {
	conn, _ := c.pool.Get()
	defer c.pool.Put(conn)

	if err := conn.Cmd("DEL", md5Uri).Err; err != nil {
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

	conn, _ := c.pool.Get()
	defer c.pool.Put(conn)

	err = conn.Cmd("MULTI").Err
	//log.Println("successfully store cacheResp ", uri)
	conn.Cmd("SET", md5Uri, b)
	conn.Cmd("EXPIRE", md5Uri, respCache.maxAge)
	err = conn.Cmd("EXEC").Err
	if err != nil {
		//log.Println(err)
		return
	}

}

func (c *ConnCachePool) Clear(d time.Duration) {

}
