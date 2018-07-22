package handler

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/Sirupsen/logrus"
	"github.com/panjf2000/goproxy/config"
	"github.com/panjf2000/goproxy/interface"
	"github.com/panjf2000/goproxy/tool"
)

var cachePool api.CachePool
var cacheLog *logrus.Logger

func init() {
	logPath := config.RuntimeViper.GetString("server.log_path")
	os.MkdirAll(logPath, os.ModePerm)
	cacheLog, _ = tool.InitLog(path.Join(logPath, "cache.log"))
}

func RegisterCachePool(c api.CachePool) {
	cachePool = c
}

//CacheHandler handles "Get" request
func (ps *ProxyServer) CacheHandler(rw http.ResponseWriter, req *http.Request) {

	var uri = req.RequestURI

	c := cachePool.Get(uri)

	if c != nil {
		if c.Verify() {
			cacheLog.WithFields(logrus.Fields{
				"request url": uri,
			}).Debug("Found cache!")
			c.WriteTo(rw)
			return
		}
		cacheLog.WithFields(logrus.Fields{
			"request url": uri,
		}).Debug("Delete cache!")
		cachePool.Delete(uri)
	}

	RmProxyHeaders(req)
	resp, err := ps.Travel.RoundTrip(req)
	if err != nil {
		http.Error(rw, err.Error(), 500)
		return
	}
	defer resp.Body.Close()

	cresp := new(http.Response)
	*cresp = *resp
	CopyResponse(cresp, resp)

	cacheLog.WithFields(logrus.Fields{
		"request url": uri,
	}).Debug("Check out this cache and then stores it if it is right!")
	go cachePool.CheckAndStore(uri, req, cresp)

	ClearHeaders(rw.Header())
	CopyHeaders(rw.Header(), resp.Header)

	rw.WriteHeader(resp.StatusCode) //写入响应状态

	nr, err := io.Copy(rw, resp.Body)
	if err != nil && err != io.EOF {
		cacheLog.WithFields(logrus.Fields{
			"client": ps.Browser,
			"error":  err,
		}).Error("occur an error when copying remote response to this client")
		return
	}
	cacheLog.WithFields(logrus.Fields{
		"response bytes": nr,
		"request url":    req.URL.Host,
	}).Info("response has been copied successfully!")
}

func CopyResponse(dest *http.Response, src *http.Response) {
	*dest = *src
	var bodyBytes []byte

	if src.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(src.Body)
	}

	// Restore the io.ReadCloser to its original state
	src.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	dest.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
}
