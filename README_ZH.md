<div align="center"><img src="https://raw.githubusercontent.com/panjf2000/logos/master/goproxy/logo.png"/></div>


[![Build Status](https://travis-ci.org/panjf2000/goproxy.svg?branch=master)](https://travis-ci.org/panjf2000/goproxy)
[![Goproxy on Sourcegraph](https://sourcegraph.com/github.com/panjf2000/goproxy/-/badge.svg)](https://sourcegraph.com/github.com/panjf2000/goproxy?badge)
[![GPL Licence](https://badges.frapsoft.com/os/gpl/gpl.svg?v=103)](https://opensource.org/licenses/GPL-3.0/)
[![Open Source Love](https://badges.frapsoft.com/os/v2/open-source.svg?v=103)](https://github.com/ellerbrock/open-source-badges/)

🇨🇳中文|[English](README.md)

## goproxy

goproxy 是使用 Go 实现的一个基本的负载均衡服务器，支持缓存（使用内存或者 Redis）；负载均衡目前支持：随机挑选一个服务器、轮询法（加权轮询）、p2c 负载均衡算法、IP HASH 模式，根据 client ip 用 hash ring 择取服务器、边界一致性哈希算法 6 种模式。另外，对转发的请求有较大的控制度，可以控制代理特定的请求，屏蔽特定的请求，甚至可以重写特定的请求。 

另外，有时候项目需要用到第三方的服务并对返回的数据进行自定义修改，调用第三方的 API，利用 goproxy 可以很容易的控制第三方 API 返回的数据并进行自定义修改。

![](https://raw.githubusercontent.com/panjf2000/illustrations/master/go/reverseproxy.png)

## 🚀 功能：

- 反向代理、负载均衡，负载策略目前支持 6 种算法：随机选取、IP HASH两种模式、轮询（Round Robin）法、加权轮询（Weight Round Robin）法、Power of Two Choices (P2C)算法、边界一致性哈希算法（Consistent Hashing with Bounded Loads）
- 支持 GET/POST/PUT/DELETE 这些 HTTP Methods，还有 HTTPS 的 CONNECT 方法
- 支持 HTTP authentication
- 支持屏蔽/过滤第三方 API 
- 支持改写 responses
- 支持内容缓存和重校验：把 response 缓存在内存或者 Redis，定期刷新，加快请求响应速度
- 灵活的 configurations 配置，支持热加载

## 🎉 使用

### 1.获取源码

```powershell
go get github.com/panjf2000/goproxy
```
**如果开启 Redis 配置，则需要额外安装 Redis。**

### 2.编译源码
```powershell
cd $GOPATH/src/github.com/panjf2000/goproxy

go build
```

### 3.运行
先配置 cfg.toml 配置文件，cfg.toml 配置文件默认存放路径为 `/etc/proxy/cfg.toml`，需要在该目录预先放置一个 cfg.toml 配置文件，一个典型的例子如下：
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

### 配置项：
#### [server]
- **port**：代理服务器的监听端口
- **reverse**：设置反向代理，值为 true 或者 false
- **proxy_pass**：反向代理目标服务器地址列表，如["127.0.0.1:80^10","127.0.0.1:88^5","127.0.0.1:8088^2","127.0.0.1:8888"]，目前支持设置服务器权重，依权重优先转发请求
- **inverse_mode**：设置负载策略，即选择转发的服务器，目前支持模式：0-随机挑选一个服务器； 1-轮询法（加权轮询）； 2-p2c负载均衡算法； 3-IP HASH 模式，根据 client ip 用 hash ring 择取服务器； 4-边界一致性哈希算法
- **auth**：开启代理认证，值为 true 或者 false
- **cache**：是否开启缓存（缓存response），值为 true 或者 false
- **cache_timeout**：redis 缓存 response 的刷新时间，以秒为单位
- **cache_type**: redis 或者 memory
- **log**：设置 log 的 level，值为 1 表示 Debug，值为 0 表示 info
- **log_path**：设置存放 log 的路径
- **user**：代理服务器的 http authentication 用户
- **http_read_timeout**：代理服务器读取 http request 的超时时间，一旦超过该时长，就会抛出异常
- **http_write_timeout**：代理服务器转发后端真实服务器时写入 http response 的超时时间，一旦超过该时长，就会抛出异常

#### [redis]
- **redis_host**：缓存模块的 redis host
- **redis_pass**：redis 密码
- **max_idle**：redis 连接池最大空闲连接数
- **idle_timeout**：空闲连接超时关闭设置
- **max_active**：连接池容量

#### [mem]

- **capacity**：缓存容量
- **cache_replacement_policy**：LRU 或者 LFU 算法

运行完go build后会生成一个执行文件，名字与项目名相同，可以直接运行：./goproxy 启动反向代理服务器。

goproxy 运行之后会监听配置文件中设置的 port 端口，然后直接访问该端口即可实现反向代理，将请求转发至proxy_pass参数中的服务器。

## 🎱 二次开发

目前该项目已实现反向代理负载均衡，支持缓存，也可以支持开发者精确控制请求，如屏蔽某些请求或者重写某些请求，甚至于对 response 进行自定义修改（定制 response 的内容），要实现精确控制 request，只需继承（不严谨的说法，因为实际上 golang 没有面向对象的概念）handlers/proxy.go 中的 ProxyServer struct，重写它的 ServeHTTP 方法，进行自定义的处理即可。

## 🙏🏻 致谢

- [httpproxy](https://github.com/sakeven/httpproxy)
- [gcache](https://github.com/bluele/gcache)
- [viper](https://github.com/spf13/viper)
- [redigo](https://github.com/gomodule/redigo)
