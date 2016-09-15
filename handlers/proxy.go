package handlers

import (
	"fmt"
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

	remote, err := url.Parse(protocol + r.Host + ":" + p.Port)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ServeHTTP(w, r)
}

//func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	protocol := r.Proto
//	portIndex := strings.Index(host, ":")
//	host := r.Host[:portIndex]
//	port := p.Port
//	method := r.Method
//	body := r.Body
//	header := r.Header
//	uri := r.RequestURI
//	setURL := fmt.Sprintf("%s://%s:%s%s", protocol, host, port, uri)
//	client := &http.Client{}
//
//
//}
