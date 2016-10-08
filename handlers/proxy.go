package handlers

import (
	_ "bufio"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/panjf2000/goproxy/cache"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

type ProxyServer struct {
	// Browser records user's name
	Travel  *http.Transport
	Browser string
}

var proxyLog *logrus.Logger

func init() {
	var filename string = "logs/proxy.log"
	proxyLog = logrus.New()
	// Log as JSON instead of the default ASCII formatter.
	proxyLog.Formatter = &logrus.TextFormatter{}

	// Output to stderr instead of stdout, could also be a file.
	if cache.CheckFileIsExist(filename) {
		f, err := os.OpenFile(filename, os.O_APPEND, 0666)
		if err != nil {
			return
		}
		proxyLog.Out = f
	} else {
		f, err := os.Create(filename)
		if err != nil {
			return
		}
		proxyLog.Out = f
	}

	// Only log the warning severity or above.
	proxyLog.Level = logrus.DebugLevel

}

// NewProxyServer returns a new proxyserver.
func NewProxyServer() *http.Server {
	if conf.Cache {
		RegisterCachePool(cache.NewCachePool(conf.RedisHost, conf.RedisPasswd, 10))
	}

	return &http.Server{
		Addr:           conf.Port,
		Handler:        &ProxyServer{Travel: &http.Transport{Proxy: http.ProxyFromEnvironment, DisableKeepAlives: true}},
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

//ServeHTTP will be automatically called by system.
//ProxyServer implements the Handler interface which need ServeHTTP.
func (goproxy *ProxyServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			//log.Debug("Panic: %v\n", err)
			fmt.Fprintf(rw, fmt.Sprintln(err))
		}
	}()
	if !goproxy.Auth(rw, req) {
		return
	}

	goproxy.ReverseHandler(req)

	if req.Method == "CONNECT" {
		goproxy.HttpsHandler(rw, req)
	} else if conf.Cache == true && req.Method == "GET" {
		goproxy.CacheHandler(rw, req)
	} else {
		goproxy.HttpHandler(rw, req)
	}
}

//HttpHandler handles http connections.
//处理普通的http请求
func (goproxy *ProxyServer) HttpHandler(rw http.ResponseWriter, req *http.Request) {
	//log.Info("%v is sending request %v %v \n", goproxy.Browser, req.Method, req.URL.Host)
	RmProxyHeaders(req)

	resp, err := goproxy.Travel.RoundTrip(req)
	if err != nil {
		//log.Error("%v", err)
		http.Error(rw, err.Error(), 500)
		return
	}
	defer resp.Body.Close()

	ClearHeaders(rw.Header())
	CopyHeaders(rw.Header(), resp.Header)

	rw.WriteHeader(resp.StatusCode) //写入响应状态

	nr, err := io.Copy(rw, resp.Body)
	if err != nil && err != io.EOF {
		proxyLog.WithFields(logrus.Fields{
			"client": goproxy.Browser,
			"error":  err,
		}).Error("occur an error when copying remote response to this client")
		return
	}
	proxyLog.WithFields(logrus.Fields{
		"response bytes": nr,
		"request url":    req.URL.Host,
	}).Info("response has been copied successfully!")
}

var HTTP_200 = []byte("HTTP/1.1 200 Connection Established\r\n\r\n")

// HttpsHandler handles any connection which need connect method.
// 处理https连接，主要用于CONNECT方法
func (goproxy *ProxyServer) HttpsHandler(rw http.ResponseWriter, req *http.Request) {
	//log.Info("%v tried to connect to %v", goproxy.Browser, req.URL.Host)

	hj, _ := rw.(http.Hijacker)
	Client, _, err := hj.Hijack() //获取客户端与代理服务器的tcp连接
	if err != nil {
		//log.Error("%v failed to get Tcp connection of \n", goproxy.Browser, req.RequestURI)
		http.Error(rw, "Failed", http.StatusBadRequest)
		return
	}

	Remote, err := net.Dial("tcp", req.URL.Host) //建立服务端和代理服务器的tcp连接
	if err != nil {
		//log.Error("%v failed to connect %v\n", goproxy.Browser, req.RequestURI)
		http.Error(rw, "Failed", http.StatusBadGateway)
		return
	}

	Client.Write(HTTP_200)

	go copyRemoteToClient(goproxy.Browser, Remote, Client)
	go copyRemoteToClient(goproxy.Browser, Client, Remote)
}

func copyRemoteToClient(User string, Remote, Client net.Conn) {
	defer func() {
		Remote.Close()
		Client.Close()
	}()

	nr, err := io.Copy(Remote, Client)
	if err != nil && err != io.EOF {
		//log.Error("%v got an error when handles CONNECT %v\n", User, err)
		proxyLog.WithFields(logrus.Fields{
			"client": User,
			"error":  err,
		}).Error("occur an error when handling CONNECT Method")
		return
	}
	//log.Info("%v transport %v bytes betwwen %v and %v.\n", User, nr, Remote.RemoteAddr(), Client.RemoteAddr())
	proxyLog.WithFields(logrus.Fields{
		"user":           User,
		"nr":             nr,
		"remote_address": Remote.RemoteAddr(),
		"client_address": Client.RemoteAddr(),
	}).Info("transport the bytes between client and remote!")
}
