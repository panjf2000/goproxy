package handlers

import (
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"github.com/panjf2000/goproxy/tool"
	"net"
)

type ProxyHandler struct {
	Mode int
	BaseHandler
}

func (p *ProxyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	httpProto := strings.Count(req.Proto, "HTTP")
	httpsProto := strings.Count(req.Proto, "HTTPS")
	protocol := "http://"
	if httpsProto == 1 && httpProto == 0 {
		protocol = "https://"
		p.Protocol = protocol
	}
	var proxyHost string
	memcacheServers := p.Host
	switch p.Mode {
	case 0:
		// 根据客户端的IP算出一个HASH值，将请求分配到集群中的某一台服务器上
		ring := tool.New(memcacheServers)
		if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			server, _ := ring.GetNode(clientIP)
			proxyHost = server
		} else {
			proxyHost = p.Host[rand.Intn(len(p.Host))]
		}
	case 1:
		// 随机选取一个负载均衡的服务器
		index := rand.Intn(len(p.Host))
		proxyHost = p.Host[index]

	}
	remote, err := url.Parse(protocol + proxyHost)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ServeHTTP(rw, req)
}
