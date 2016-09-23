package cache

import (
    "crypto/md5"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"
    "github.com/garyburd/redigo/redis"
    "github.com/panjf2000/goproxy/interface"
)

func MD5Uri(uri string) string {
    return fmt.Sprintf("%x", md5.Sum([]byte(uri)))
}

type CachePool struct {
    pool *redis.Pool
}

func NewCachePool(address string, password string) *CachePool{
    pool := &redis.Pool{
        MaxIdle:     5,
        IdleTimeout: 1 * time.Hour,
        Dial: func() (redis.Conn, error) {
            c, err := redis.Dial("tcp", address)
            if err != nil {
                return nil, err
            }

            if password != "" {
                if _, err = c.Do("AUTH", password); err != nil {
                    c.Close()
                    return nil, err
                }
            }

            return c, nil
        },
    }

    c := pool.Get()
    defer c.Close()

    _, err := c.Do("PING")
    if err != nil {
        panic("Fail to connect to redis server")
    }
    log.Println("yes to redis")
    return &CachePool{
        pool: pool,
    }

}

func (c *CachePool) Get(uri string) api.Cache {
    log.Println("get cahche of ", uri)
    if cache := c.get(MD5Uri(uri)); cache != nil {
        //log.Println(*cache)
        return cache
    }
    return nil
}

func (c *CachePool) get(md5Uri string) *Cache {
    conn := c.pool.Get()
    defer conn.Close()

    b, err := redis.Bytes(conn.Do("GET", md5Uri))
    if err != nil || len(b) == 0 {
        log.Println(err)
        return nil
    }
    log.Println(string(b))
    cache := new(Cache)
    json.Unmarshal(b, &cache)
    return cache
}

func (c *CachePool) Delete(uri string) {
    c.delete(MD5Uri(uri))
}

func (c *CachePool) delete(md5Uri string) {
    conn := c.pool.Get()
    defer conn.Close()

    _, err := conn.Do("DEL", md5Uri)

    if err != nil {
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

    log.Println("store cache ", uri)

    md5Uri := MD5Uri(uri)
    b, err := json.Marshal(cache)
    if err != nil {
        log.Println(err)
        return
    }

    conn := c.pool.Get()
    defer conn.Close()

    conn.Send("MULTI")
    conn.Send("SET", md5Uri, b)
    conn.Send("EXPIRE", md5Uri, cache.maxAge)
    _, err = conn.Do("EXEC")
    if err != nil {
        log.Println(err)
        return
    }
    log.Println("successfully store cache ", uri)

}

func (c *CachePool) Clear(d time.Duration) {

}
