/*
@version: 1.0
@author: allanpan
@license:  Apache Licence
@contact: panjf2000@gmail.com  
@site: 
@file: redisclient.go
@time: 2017/3/20 20:29
@tag: 1,2,3
@todo: ...
*/
package tool

import (
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/mediocregopher/radix.v2/sentinel"
	"github.com/panjf2000/goproxy/config"
	"time"
)

var RedisSentinel *sentinel.Client

func init() {
	for _, address := range config.RedisSentinel[config.ENV] {
		client, err := sentinel.NewClient("tcp", address, 100, config.RedisSentinelName)
		if err == nil {
			RedisSentinel = client
			break
		}

	}

	// keep redis alive
	go func() {
		for {
			time.Sleep(1 * time.Second)
			conn, err := RedisSentinel.GetMaster(config.RedisSentinelName)
			if err != nil {
				// log if you want
				continue
			}
			conn.Cmd("PING")
			RedisSentinel.PutMaster(config.RedisSentinelName, conn)
		}
	}()

}

func NewRedisClient() (*redis.Client, error) {
	redisClient, err := RedisSentinel.GetMaster(config.RedisSentinelName)
	if err := redisClient.Cmd("AUTH", config.RedisSentinelPass).Err; err != nil {
		return nil, err
	}
	if err := redisClient.Cmd("PING").Err; err != nil && err.Error() == "EOF" {
		return NewRedisClient()
	}
	return redisClient, err
}
