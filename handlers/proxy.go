package handlers

import (
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type ProxyHandler struct {
	BaseHandler
}

func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	httpProto := strings.Count(r.Proto, "HTTP")
	httpsProto := strings.Count(r.Proto, "HTTPS")
	protocol := "http://"
	if httpsProto == 1 && httpProto == 0 {
		protocol = "https://"
		p.Protocol = protocol
	}
	// 随机选取一个负载均衡的服务器
	index := rand.Intn(len(p.Host))
	proxyHost := p.Host[index]
	remote, err := url.Parse(p.Protocol + proxyHost)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ServeHTTP(w, r)
}
