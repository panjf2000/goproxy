# goproxy
# 1.获取源码
通过svn checkout本项目的源码，源码路径为：https://github.com/panjf2000/goproxy.git

# 2.安装项目依赖的golang库：
logrus（一个开源的高性能golang日志库）；
radix（一个Redis 官方推荐的使用golang实现的redis client，轻量级和实现优雅）
logrus安装：go get github.com/Sirupsen/logrus
radix安装：go get github.com/mediocregopher/radix.v2/…

# 3.编译源码
cd $GOPATH/src/
go build

# 4.运行
先配置config.json配置文件，一个典型的例子如下：
```
{
“port”: “:8080”,
“reverse”: true,
“proxy_pass”: [
“127.0.0.1:80”
],
“mode”: 0,
“auth”: false,
“cache”: true,
“redis_host”: “127.0.0.1:6379”,
“redis_passwd”: “redis_secret”,
“cache_timeout”: 60,
“log”: 1,
“admin”: {
“Admin”: “root”
},
“user”: {
“agent”: “proxy”
}
}
```

config释义：
port：代理服务器的监听端口
reverse：设置反向代理，值为true或者false
proxy_pass：反向代理目标服务器地址列表，如[“127.0.0.1:80”,“127.0.0.1:8080”]
mode：设置负载策略，即选择转发的服务器，目前支持两种模式：1.随机挑选一个服务器 2.IP HASH模式，根据client ip用hash ring择取服务器
auth：开启代理认证，值为true或者false
cache：是否开启缓存（缓存response），值为true或者false
redis_host：缓存模块的redis host
redis_passwd：redis密码
cache_timeout：缓存更新时间，单位分钟
log：设置打log的level,1时level为Debug，0时为info
user：代理服务器的验证用户

运行完go build后会生成一个执行文件，名字与项目名相同，可以直接运行：./goproxy
运行组件后，proxy server监听配置文件中设置的port端口，然后直接访问该端口即可实现反向代理，将请求转发至proxy_pass参数中的服务器

**PS:这个项目中的模块引用路径还是我本机上的路径，也就是我的github路径，编译源码前请将源码中的引用路径修改成你自己机器上的路径。**

# 二次开发
>目前该项目已实现反向代理负载均衡，支持缓存，也可以支持开发者精确控制请求，如屏蔽某些请求或者重写某些请求，甚至于对response进行自定义修改（定制>response的内容），要实现精确控制request，只需继承handlers/proxy.go中的ProxyServer struct，重写它的ServeHTTP方法，进行自定义的处理即可。
