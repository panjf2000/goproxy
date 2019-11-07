package cache

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/bluele/gcache"
	api "github.com/panjf2000/goproxy/interface"
	"github.com/panjf2000/goproxy/tool"
)

type CacheReplacementPolicy int

const (
	LRU = iota
	LFU
)

type MemConnCachePool struct {
	cache gcache.Cache
}

func NewMemCachePool(cap int, crp CacheReplacementPolicy) *MemConnCachePool {
	var gc gcache.Cache
	switch crp {
	case LFU:
		gc = gcache.New(cap).LFU().Build()
	default:
		gc = gcache.New(cap).LRU().Build()

	}
	return &MemConnCachePool{cache: gc}

}

func (c *MemConnCachePool) Get(uri string) api.Cache {
	if respCache := c.get(tool.MD5Uri(uri)); respCache != nil {
		return respCache
	}
	return nil
}

func (c *MemConnCachePool) get(md5Uri string) *HttpCache {
	value, err := c.cache.Get(md5Uri)
	if err != nil {
		return nil
	}
	b := value.([]byte)
	if len(b) == 0 {
		return nil
	}
	respCache := new(HttpCache)
	_ = json.Unmarshal(b, respCache)

	return respCache
}

func (c *MemConnCachePool) Delete(uri string) {
	c.delete(tool.MD5Uri(uri))
}

func (c *MemConnCachePool) delete(md5Uri string) {
	c.cache.Remove(md5Uri)
	return
}

func (c *MemConnCachePool) CheckAndStore(uri string, req *http.Request, resp *http.Response) {
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

	if err := c.cache.SetWithExpire(md5Uri, b, time.Duration(respCache.maxAge)*time.Second); err != nil {
		return
	}

}

//func (c *MemConnCachePool) Clear(d time.Duration) {
//
//}
