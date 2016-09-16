package handlers

import (
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
	}

	remote, err := url.Parse(protocol + r.Host[:strings.Index(r.Host, ":") + 1] + p.Port)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ServeHTTP(w, r)
}
