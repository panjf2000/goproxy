package handlers

import (
	"github.com/panjf2000/goproxy/tool"
	"math/rand"
	"net"
	"net/http"
)

//ReverseHandler handles request for reverse proxy.
//处理反向代理请求
func (goproxy *ProxyServer) ReverseHandler(req *http.Request) {
	if conf.Reverse == true {
		//用于反向代理
		goproxy.reverseHandler(req)
	}
}

//ReverseHandler handles request for reverse proxy.
//处理反向代理请求
func (goproxy *ProxyServer) reverseHandler(req *http.Request) {
	var proxyHost string
	memcacheServers := conf.ProxyPass
	switch conf.Mode {
	case 0:
		// 根据客户端的IP算出一个HASH值，将请求分配到集群中的某一台服务器上
		ring := tool.New(memcacheServers)
		if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			server, _ := ring.GetNode(clientIP)
			if tool.IsHost(server) {
				proxyHost = server
			} else {
				fallthrough
			}
		} else {
			proxyHost = memcacheServers[rand.Intn(len(memcacheServers))]
		}
	case 1:
		// 随机选取一个负载均衡的服务器
		index := rand.Intn(len(memcacheServers))
		proxyHost = memcacheServers[index]

	}
	req.Host = proxyHost
	req.URL.Host = req.Host
	req.URL.Scheme = "http"
}
