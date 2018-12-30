<div align="center"><img src="goproxy_logo.png"/></div>


[![Build Status](https://travis-ci.org/panjf2000/goproxy.svg?branch=master)](https://travis-ci.org/panjf2000/goproxy)
[![GPL Licence](https://badges.frapsoft.com/os/gpl/gpl.svg?v=103)](https://opensource.org/licenses/GPL-3.0/)
[![Open Source Love](https://badges.frapsoft.com/os/v2/open-source.svg?v=103)](https://github.com/ellerbrock/open-source-badges/)


# Changelog in 16/04/2018
## Change redis client
Change the redis client in goproxy from radix.v2 to redigo.

Reason：

>There is a potential issue in the connection pool of radix.v2 which generates new connection to redis even though the current number of connections exceeds the maximum size of radix.v2 pool and this implementation may led to the redis cluster collapse.
>
>I had tried to submit a PR to radix.v2 expecting to fix it but the author from radix.v2 rejected that PR saying that logic in radix.v2 pool was intended..., well, he is the author of radix.v2 so he got the right to determine what radix.v2 should be. However, based on the conflicting ideas from us, I'm afraid I have to change the redis client in goproxy from radix.v2 to redigo whose connection pool will restrict the size of connections severely.

# Changelog in 11/02/2018
## Optimization in configurations management of proxy server

>Managing configurations in proxy server with [viper](https://github.com/spf13/viper) which supports eager Loading and means it will take effect immediately right after you update the configuration file. 

# Changelog in 22/07/2017
## Add 4 new algorithms in goproxy for load balancing:

* Round Robin Algorithm
* Weight Round Robin Algorithm
* Power of Two Choices (P2C) Algorithm
* Consistent Hashing with Bounded Loads Algorithm


# goproxy
>goproxy is a load-balancing, reverse-proxy server implemented in go supporting cache( by redis); As a reverse-proxy server, it supports 4 algorithms: Randomized Algorithm, Weight Round Robin Algorithm, Power of Two Choices (P2C) Algorithm, IP Hash Algorithm, Consistent Hashing with Bounded Loads Algorithm, besides, goproxy can dominate the http requests: filtering and blocking specific requests and even rewriting them.
>
>Sometimes your program needs to call some third party API and wants to customize the responses from it, in that case, goproxy will be your great choice.

# Features：

## 1.Revers-proxy, load-balancing, 2 algorithms for load-balancing in goproxy

- GET/POST/PUT/DELETE methods in http and CONNECT method in https are supported in goproxy
- Http authentication was also supported
- Weight can be assigned to every single back-end server

## 2.Content forwarding：
- Filtering and blocking specific http requests and even rewriting them
- Customizing responses from third-party API

## 3.Responses can be cached in redis to speed up the responding and the expired time of caches is configurable

## 4.Configurations are stored in a json file which is convenient for users

# How to use poroxy
## 1.Get source code

>* Clone source of goproxy from github, [goproxy](https://github.com/panjf2000/goproxy)
>* git clone https://github.com/panjf2000/goproxy.git

## 2.Get those dependencies needed from goproxy：

- logrus (structured, pluggable logging for Go)
- ~~radix.v2 (lightweight redis client for Go)~~
- redigo（Go client for Redis）
- cron（a *cron* library for go）

>Install logrus：go get github.com/Sirupsen/logrus

>~~Install radix.v2：go get github.com/mediocregopher/radix.v2/…~~

>Install redigo：go get github.com/gomodule/redigo/redis

**Besides, you also need a redis to support caching responses in goproxy.**

## 3.Compile the source code
1. cd $GOPATH/src/
2. go build

## 4.Run
goproxy uses cfg.toml as its configurations file which is located to the path: /etc/proxy/cfg.toml of your server, you should create a cfg.toml in there previously, here is a typical example:

```
[server]
port = ":8080"
reverse = true
proxy_pass = ["127.0.0.1:6000", "127.0.0.1:7000", "127.0.0.1:8000", "127.0.0.1:9000"]
inverse_mode = 2
auth = false
cache = true
cache_timeout = 60
log = 1
log_path = "./logs"
user = { agent = "proxy" }
http_read_timeout = 10
http_write_timeout = 10

[redis]
redis_host = "127.0.0.1:6379"
redis_pass = "redis_pass"
max_idle = 5
idle_timeout = 10
max_active = 10

```

### config释义：
#### [server]
- port：the port goroxy will listen to
- reverse：enable the reverse-proxy feature or not
- proxy_pass：back-end servers that actually provide services, like ["127.0.0.1:80^10","127.0.0.1:88^5","127.0.0.1:8088^2","127.0.0.1:8888"], weight can be assigned to every single server
- inverse_mode：load-balancing algoritms：0 for Randomized Algorithm； 1 for Weight Round Robin Algorithm； 2 for Power of Two Choices (P2C) Algorithm； 3 for IP Hash Algorithm； 4 for Consistent Hashing with Bounded Loads Algorithm
- auth：enable http authentication or not
- cache：enable responses caching or not
- cache_timeout：expired time of responses caching, in minutes
- log：log level, 1 for Debug，0 for info
- log_path：the path of log files
- user：user name from http authentication
- http_read_timeout：duration for waiting response from the back-end server, if goproxy don't get the response after this duration, it will throw an exception
- http_write_timeout：duration for back-end server writing response to goproxy, if back-end server takes a longer time than this duration to write its response into goproxy, goproxy will throw an exception

#### [redis]
- redis_host：redis host
- redis_pass：redis password
- max_idle：the maximum idle connections of redis connection pool
- idle_timeout：duration for idle redis connection to close
- max_active：maximum size of redis connection pool

You will get a binary file after executing the `go build` command named goproxy as the same of project name which can be run directly to start a proxy server.

The started proxy server will listen the port set in cfg.toml file and you can just forward your http requests to the back-end servers set in cfg.toml by going through that port in goproxy.

# Secondary development
>Up to present, goproxy has implemented all basic functionalities like reverse-proxy, load-blancing, http caching, http requests controlling, etc and if you want to customize the responses more accurately, you can implement a new handler by inheriting (not a strict statement as there is no OO in golang) from the ProxyServer struct located in handlers/proxy.go and overriding its method named ServeHTTP, then you are allowed to write your own logic into it.