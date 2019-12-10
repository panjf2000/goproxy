<div align="center"><img src="https://raw.githubusercontent.com/panjf2000/logos/master/goproxy/logo.png"/></div>


[![Build Status](https://travis-ci.org/panjf2000/goproxy.svg?branch=master)](https://travis-ci.org/panjf2000/goproxy)
[![Goproxy on Sourcegraph](https://sourcegraph.com/github.com/panjf2000/goproxy/-/badge.svg)](https://sourcegraph.com/github.com/panjf2000/goproxy?badge)
[![GPL Licence](https://badges.frapsoft.com/os/gpl/gpl.svg?v=103)](https://opensource.org/licenses/GPL-3.0/)
[![Open Source Love](https://badges.frapsoft.com/os/v2/open-source.svg?v=103)](https://github.com/ellerbrock/open-source-badges/)

Engilish|[🇨🇳中文](README_ZH.md)

## goproxy

goproxy is a load-balancing, reverse-proxy server implemented in go, supporting cache( in memory or Redis). As a load-balancing server, it supports 4 algorithms: Randomized Algorithm, Weight Round Robin Algorithm, Power of Two Choices (P2C) Algorithm, IP Hash Algorithm, Consistent Hashing with Bounded Loads Algorithm, besides, goproxy can dominate the http requests: filtering and blocking specific requests and even rewriting them.

Sometimes your program needs to call some third party API and wants to customize the responses from it, in that case, goproxy will be your great choice.

![](https://raw.githubusercontent.com/panjf2000/illustrations/master/go/reverseproxy.png)

## 🚀 Features：

- Supporting reverse-proxy, 6 load-balancing algorithms in goproxy: Random, IP Hash, Round Robin, Weight Round Robin, Power of Two Choices (P2C), Consistent Hashing with Bounded Loads
- Supporting GET/POST/PUT/DELETE Methods in http and CONNECT method in https in goproxy
- Supporting HTTP authentication
- Filtering and blocking specific http requests and even rewriting them in goproxy
- Customizing responses from third-party API
- Cache support with memory or Redis to speed up the responding and the expired time of caches is configurable
- Flexible and eager-loading configurations

## 🎉 How to use goproxy

### 1.Get source code

```powershell
go get github.com/panjf2000/goproxy
```

**Besides, you also need a Redis to support caching responses if you enable Redis config in goproxy.**

### 2.Compile the source code

```powershell
cd $GOPATH/src/github.com/panjf2000/goproxy

go build
```

### 3.Run

goproxy uses cfg.toml as its configurations file which is located in `/etc/proxy/cfg.toml` of your server, you should create a cfg.toml in there previously, here is a typical example:

```toml
# toml file for goproxy

title = "TOML config for goproxy"

[server]
port = ":8080"
reverse = true
proxy_pass = ["127.0.0.1:6000"]
# 0 - random, 1 - loop, 2 - power of two choices(p2c), 3 - hash, 4 - consistent hashing
inverse_mode = 2
auth = false
cache = true
cache_timeout = 60
cache_type = "redis"
log = 1
log_path = "./logs"
user = { agent = "proxy" }
http_read_timeout = 10
http_write_timeout = 10

[redis]
redis_host = "localhost:6379"
redis_pass = ""
max_idle = 5
idle_timeout = 10
max_active = 10

[mem]
capacity = 1000
cache_replacement_policy = "LRU"
```

### Configurations：
#### [server]
- **port**：the port goroxy will listen to
- **reverse**：enable the reverse-proxy feature or not
- **proxy_pass**：back-end servers that actually provide services, like ["127.0.0.1:80^10","127.0.0.1:88^5","127.0.0.1:8088^2","127.0.0.1:8888"], weight can be assigned to every single server
- **inverse_mode**：load-balancing algoritms：0 for Randomized Algorithm； 1 for Weight Round Robin Algorithm； 2 for Power of Two Choices (P2C) Algorithm； 3 for IP Hash Algorithm based on hash ring； 4 for Consistent Hashing with Bounded Loads Algorithm
- **auth**：enable http authentication or not
- **cache**：enable responses caching or not
- **cache_timeout**：expired time of responses caching, in seconds
- **cache_type**: redis or memory
- **log**：log level, 1 for Debug，0 for info
- **log_path**：the path of log files
- **user**：user name from http authentication
- **http_read_timeout**：duration for waiting response from the back-end server, if goproxy don't get the response after this duration, it will throw an exception
- **http_write_timeout**：duration for back-end server writing response to goproxy, if back-end server takes a longer time than this duration to write its response into goproxy, goproxy will throw an exception

#### [redis]
- **redis_host**：redis host
- **redis_pass**：redis password
- **max_idle**：the maximum idle connections of redis connection pool
- **idle_timeout**：duration for idle redis connection to close
- **max_active**：maximum size of redis connection pool

#### [mem]

- **capacity**: cache capacity of items
- **cache_replacement_policy**: LRU or LFU

There should be a binary named `goproxy` as the same of project name after executing the `go build` command and that binary can be run directly to start a goproxy server.

The running goproxy server listens in the port set in cfg.toml and it will forward your http requests to the back-end servers set in cfg.toml by going through that port in goproxy.

## 🎱 Secondary development

Up to present, goproxy has implemented all basic functionalities like reverse-proxy, load-blancing, http caching, http requests controlling, etc and if you want to customize the responses more accurately, you can implement a new handler by inheriting (not a strict statement as there is no OO in golang) from the ProxyServer struct located in handlers/proxy.go and overriding its method named ServeHTTP, then you are allowed to write your own logic into it.

## 🙏🏻 Thanks

- [httpproxy](https://github.com/sakeven/httpproxy)
- [gcache](https://github.com/bluele/gcache)
- [viper](https://github.com/spf13/viper)
- [redigo](https://github.com/gomodule/redigo)