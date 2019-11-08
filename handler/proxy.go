package handler

import (
	_ "bufio"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/panjf2000/goproxy/cache"
	"github.com/panjf2000/goproxy/config"
)

type ProxyServer struct {
	// Browser records user's name
	Travel  *http.Transport
	Browser string
}

// NewProxyServer returns a new proxy server.
func NewProxyServer() *http.Server {
	if config.RuntimeViper.GetBool("server.cache") {
		var cachePoolType cache.CachePoolType
		if config.RuntimeViper.GetString("server.cache_type") == "redis" {
			cachePoolType = cache.Redis
		}
		RegisterCachePool(cache.NewCachePool(cachePoolType))
	}

	return &http.Server{
		Addr:           config.RuntimeViper.GetString("server.port"),
		Handler:        &ProxyServer{Travel: &http.Transport{Proxy: http.ProxyFromEnvironment, DisableKeepAlives: false}},
		ReadTimeout:    time.Duration(config.RuntimeViper.GetInt("server.http_read_timeout")) * time.Second,
		WriteTimeout:   time.Duration(config.RuntimeViper.GetInt("server.http_write_timeout")) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

//ServeHTTP will be automatically called by system.
//ProxyServer implements the Handler interface which need ServeHTTP.
func (ps *ProxyServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
		}
	}()
	if !ps.Auth(rw, req) {
		return
	}

	ps.LoadBalancing(req)
	defer ps.Done(req)

	if req.Method == "CONNECT" {
		ps.HttpsHandler(rw, req)
	} else if req.Method == "GET" && config.RuntimeViper.GetBool("server.cache") {
		ps.CacheHandler(rw, req)
	} else {
		ps.HttpHandler(rw, req)
	}
}

//HttpHandler handles http connections.
func (ps *ProxyServer) HttpHandler(rw http.ResponseWriter, req *http.Request) {
	RmProxyHeaders(req)

	resp, err := ps.Travel.RoundTrip(req)
	if err != nil {
		http.Error(rw, err.Error(), 500)
		return
	}
	defer resp.Body.Close()

	ClearHeaders(rw.Header())
	CopyHeaders(rw.Header(), resp.Header)

	rw.WriteHeader(resp.StatusCode) // writes the response status.

	_, err = io.Copy(rw, resp.Body)
	if err != nil && err != io.EOF {
		return
	}
}


// HttpsHandler handles any connection which needs "connect" method.
func (ps *ProxyServer) HttpsHandler(rw http.ResponseWriter, req *http.Request) {
	hj, _ := rw.(http.Hijacker)
	Client, _, err := hj.Hijack() // gets the tcp connection between client and server.
	if err != nil {
		http.Error(rw, "Failed", http.StatusBadRequest)
		return
	}

	Remote, err := net.Dial("tcp", req.URL.Host) // establishes the tcp connection between the client and server.
	if err != nil {
		http.Error(rw, "Failed", http.StatusBadGateway)
		return
	}

	_, _ = Client.Write(HTTP200)

	go copyRemoteToClient(ps.Browser, Remote, Client)
	go copyRemoteToClient(ps.Browser, Client, Remote)
}

func copyRemoteToClient(User string, Remote, Client net.Conn) {
	defer func() {
		_ = Remote.Close()
		_ = Client.Close()
	}()

	_, err := io.Copy(Remote, Client)
	if err != nil && err != io.EOF {
		return
	}
}
