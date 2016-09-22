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

type CacheHolder struct {
    pool *redis.Pool
}

func NewCacheHolder(address string, password string) *CacheHolder {
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
    return &CacheHolder{
        pool: pool,
    }

}

func (c *CacheHolder) Get(uri string) api.Cache {
    log.Println("get cahche of ", uri)
    if cache := c.get(MD5Uri(uri)); cache != nil {
        //log.Println(*cache)
        return cache
    }
    return nil
}

func (c *CacheHolder) get(md5Uri string) *Cache {
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

func (c *CacheHolder) Delete(uri string) {
    c.delete(MD5Uri(uri))
}

func (c *CacheHolder) delete(md5Uri string) {
    conn := c.pool.Get()
    defer conn.Close()

    _, err := conn.Do("DEL", md5Uri)

    if err != nil {
        return
    }

    return
}

func (c *CacheHolder) CheckAndStore(uri string, resp *http.Response) {
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

func (c *CacheHolder) Clear(d time.Duration) {

}
