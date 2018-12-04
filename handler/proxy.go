package handler

import (
	_ "bufio"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/panjf2000/goproxy/cache"
	"github.com/panjf2000/goproxy/config"
	"github.com/panjf2000/goproxy/tool"
)

type ProxyServer struct {
	// Browser records user's name
	Travel  *http.Transport
	Browser string
}

var proxyLog *logrus.Logger

func init() {
	logPath := config.RuntimeViper.GetString("server.log_path")
	os.MkdirAll(logPath, os.ModePerm)
	proxyLog, _ = tool.InitLog(path.Join(logPath, "proxy.log"))

}

// NewProxyServer returns a new proxy server.
func NewProxyServer() *http.Server {
	if config.RuntimeViper.GetBool("server.cache") {
		RegisterCachePool(cache.NewCachePool(config.RuntimeViper.GetString("redis.redis_host"),
			config.RuntimeViper.GetString("redis.redis_pass"), config.RuntimeViper.GetInt("redis.idle_timeout"),
			config.RuntimeViper.GetInt("redis.max_active"), config.RuntimeViper.GetInt("redis.max_idle")))
	}

	return &http.Server{
		Addr:           config.RuntimeViper.GetString("server.port"),
		Handler:        &ProxyServer{Travel: &http.Transport{Proxy: http.ProxyFromEnvironment, DisableKeepAlives: false}},
		ReadTimeout:    time.Duration(config.RuntimeViper.GetInt("server.http_read_timeout")) * time.Second,
		WriteTimeout:   time.Duration(config.RuntimeViper.GetInt("server.http_write_timeout")) * time.Second,
		MaxHeaderBytes: 1 << 20,
		ErrorLog:       log.New(proxyLog.Out, "[ERROR]", log.LstdFlags),
	}
}

//ServeHTTP will be automatically called by system.
//ProxyServer implements the Handler interface which need ServeHTTP.
func (ps *ProxyServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			proxyLog.WithFields(logrus.Fields{
				"panic": err,
			}).Panic("Call a panic!")
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
	proxyLog.WithFields(logrus.Fields{
		"request user":   ps.Browser,
		"request method": req.Method,
		"request url":    req.URL.Host,
	}).Info("request's detail !")
	RmProxyHeaders(req)

	resp, err := ps.Travel.RoundTrip(req)
	if err != nil {
		proxyLog.WithFields(logrus.Fields{
			"error": err,
		}).Error("occur an error!")
		http.Error(rw, err.Error(), 500)
		return
	}
	defer resp.Body.Close()

	ClearHeaders(rw.Header())
	CopyHeaders(rw.Header(), resp.Header)

	rw.WriteHeader(resp.StatusCode) // write the response status.

	nr, err := io.Copy(rw, resp.Body)
	if err != nil && err != io.EOF {
		proxyLog.WithFields(logrus.Fields{
			"client": ps.Browser,
			"error":  err,
		}).Error("occur an error when copying remote response to this client")
		return
	}
	proxyLog.WithFields(logrus.Fields{
		"response bytes": nr,
		"request url":    req.URL.Host,
	}).Info("response has been copied successfully!")
}

var HTTP200 = []byte("HTTP/1.1 200 Connection Established\r\n\r\n")

// HttpsHandler handles any connection which needs "connect" method.
func (ps *ProxyServer) HttpsHandler(rw http.ResponseWriter, req *http.Request) {
	proxyLog.WithFields(logrus.Fields{
		"user": ps.Browser,
		"host": req.URL.Host,
	}).Info("http user tried to connect host!")

	hj, _ := rw.(http.Hijacker)
	Client, _, err := hj.Hijack() // get the tcp connection between client and server.
	if err != nil {
		proxyLog.WithFields(logrus.Fields{
			"user":        ps.Browser,
			"request uri": req.RequestURI,
		}).Error("http user failed to get tcp connection!")
		http.Error(rw, "Failed", http.StatusBadRequest)
		return
	}

	Remote, err := net.Dial("tcp", req.URL.Host) // establish the tcp connection between the client and server.
	if err != nil {
		proxyLog.WithFields(logrus.Fields{
			"user":        ps.Browser,
			"request uri": req.RequestURI,
		}).Error("http user failed to connect this uri!")
		http.Error(rw, "Failed", http.StatusBadGateway)
		return
	}

	Client.Write(HTTP200)

	go copyRemoteToClient(ps.Browser, Remote, Client)
	go copyRemoteToClient(ps.Browser, Client, Remote)
}

func copyRemoteToClient(User string, Remote, Client net.Conn) {
	defer func() {
		Remote.Close()
		Client.Close()
	}()

	nr, err := io.Copy(Remote, Client)
	if err != nil && err != io.EOF {
		proxyLog.WithFields(logrus.Fields{
			"client": User,
			"error":  err,
		}).Error("occur an error when handling CONNECT Method")
		return
	}
	proxyLog.WithFields(logrus.Fields{
		"user":           User,
		"nr":             nr,
		"remote_address": Remote.RemoteAddr(),
		"client_address": Client.RemoteAddr(),
	}).Info("transport the bytes between client and remote!")
}
