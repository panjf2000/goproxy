package cache

import (
	"github.com/panjf2000/goproxy/config"
	api "github.com/panjf2000/goproxy/interface"
)

type CachePoolType int

const (
	Mem = iota
	Redis
)

func NewCachePool(cachePoolType CachePoolType) api.CachePool {
	switch cachePoolType {
	case Redis:
		return NewRedisCachePool(config.RuntimeViper.GetString("redis.redis_host"),
			config.RuntimeViper.GetString("redis.redis_pass"), config.RuntimeViper.GetInt("redis.idle_timeout"),
			config.RuntimeViper.GetInt("redis.max_active"), config.RuntimeViper.GetInt("redis.max_idle"))
	default:
		var crp CacheReplacementPolicy
		if config.RuntimeViper.GetString("mem.cache_replacement_policy") == "LFU" {
			crp = LFU
		}
		return NewMemCachePool(config.RuntimeViper.GetInt("mem.capacity"), crp)
	}
}
