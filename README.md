<div align="center"><img src="goproxy_logo.png"/></div>


[![Build Status](https://travis-ci.org/panjf2000/goproxy.svg?branch=master)](https://travis-ci.org/panjf2000/goproxy)
[![GPL Licence](https://badges.frapsoft.com/os/gpl/gpl.svg?v=103)](https://opensource.org/licenses/GPL-3.0/)
[![Open Source Love](https://badges.frapsoft.com/os/v2/open-source.svg?v=103)](https://github.com/ellerbrock/open-source-badges/)


# 2018.04.16更新
## 更换redis client
redis客户端由原来的radix.v2库更换为redigo库
更换理由：
>radix.v2库的连接池有一个潜在问题是，如果同时初始化大量的连接，即使超过了pool的size，
radix.v2依然会不断申请新的redis连接，如果在极端情况下，大量的连接建立可能会导致
redis server的崩溃，本人向radix.v2的作者提交了一个pr，但作者并不接受，
且说是设计如此...无力吐槽，但因为他是作者且不同的理念有时候无法调和，
所以本项目只能换redis库，故迁移到有连接池保护的redigo。

# 2018.02.11更新
## 优化server的config管理
>使用viper库和toml文件来管理server的配置信息，并且实现热加载，修改配置文件后实时生效，无需重启server。

# 2017.07.22项目更新
## 新增4种负载均衡算法：
* 轮询（Round Robin）法
* 加权轮询（Weight Round Robin）法
* Power of Two Choices (P2C)算法
* 边界一致性哈希算法（Consistent Hashing with Bounded Loads）


# goproxy
>goproxy是使用golang实现的一个基本的负载均衡服务器，支持缓存（使用redis）；反向代理，目前支持随机分发和IP HASH两种模式，另外，对转发的请求有较大的控制度，可以控制代理特定的请求，屏蔽特定的请求，甚至可以重写特定的请求。 另外，有时候项目需要用到第三方的服务并对返回的数据进行自定义修改，调用第三方的API，利用proxy server可以很容易的控制第三方API返回的数据并进行自定义修改。

# 项目功能：

## 1.反向代理、负载均衡，负载策略目前支持随机选取和IP HASH两种模式；
- 支持GET/POST/PUT/DELETE这些Method，还有https的CONNECT方法
- 支持http authentication
- 负载策略支持预设权重，依权重优先转发请求

## 2.内容转发：
- 可以控制代理特定的请求，屏蔽特定的请求，甚至可以重写特定的请求,
- 控制第三方API返回的数据并进行自定义修改

## 3.支持内容缓存和重校验，支持把response缓存在redis，定时刷新，加快请求响应速度。

## 4.通过config文件实现对server的配置

# 系统使用
## 1.获取源码
>* 通过github获取项目的[源码](https://github.com/panjf2000/goproxy)，路径为：https://github.com/panjf2000/goproxy
>* 获取：git clone https://github.com/panjf2000/goproxy.git

## 2.安装项目依赖的golang库：
- logrus（一个开源的高性能golang日志库）;
- ~~radix.v2（一个Redis 官方推荐的使用golang实现的redis client，轻量级、实现优雅）;~~
- redigo（redis官方推荐client）
- cron（golang实现的一个crontab）

>logrus安装：go get github.com/Sirupsen/logrus

>~~radix.v2安装：go get github.com/mediocregopher/radix.v2/…~~

>redigo安装：go get github.com/gomodule/redigo/redis

**另外，该项目需要redis数据库的支持，所以要有一个redis环境**

## 3.编译源码
1. cd $GOPATH/src/
2. go build

## 4.运行
先配置cfg.toml 配置文件，cfg.toml配置文件默认存放路径为/etc/proxy/cfg.toml,请在该目录预先置放一个cfg.toml配置文件，一个典型的例子如下：
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

[redis]
redis_host = "127.0.0.1:6379"
redis_pass = "redis_pass"
max_idle = 5
idle_timeout = 10
max_active = 10

```

### config释义：
#### [server]
- port：代理服务器的监听端口
- reverse：设置反向代理，值为true或者false
- proxy_pass：反向代理目标服务器地址列表，如["127.0.0.1:80^10","127.0.0.1:88^5","127.0.0.1:8088^2","127.0.0.1:8888"]，目前支持设置服务器权重，依权重优先转发请求
- inverse_mode：设置负载策略，即选择转发的服务器，目前支持模式：0-随机挑选一个服务器； 1-轮询法（加权轮询）； 2-p2c负载均衡算法； 3-IP HASH模式，根据client ip用hash ring择取服务器； 4-边界一致性哈希算法
- auth：开启代理认证，值为true或者false
- cache：是否开启缓存（缓存response），值为true或者false
- cache_timeout：redis缓存response的刷新时间，以分钟为单位
- log：设置打log的level,1时level为Debug，0时为info
- log_path：设置存放log的路径
- user：代理服务器的http authentication 用户

#### [redis]
- redis_host：缓存模块的redis host
- redis_pass：redis密码
- max_idle：redis连接池最大空闲连接数
- idle_timeout：空闲连接超时关闭设置
- max_active：连接池容量

  
  
运行完go build后会生成一个执行文件，名字与项目名相同，可以直接运行：./goproxy
运行组件后，proxy server监听配置文件中设置的port端口，然后直接访问该端口即可实现反向代理，将请求转发至proxy_pass参数中的服务器

**PS:这个项目中的模块引用路径还是我本机上的路径，也就是我的github路径，编译源码前请将源码中的引用路径修改成你自己机器上的路径。**

# 二次开发
>目前该项目已实现反向代理负载均衡，支持缓存，也可以支持开发者精确控制请求，如屏蔽某些请求或者重写某些请求，甚至于对response进行自定义修改（定制response的内容），要实现精确控制request，只需继承handlers/proxy.go中的ProxyServer struct，重写它的ServeHTTP方法，进行自定义的处理即可。
