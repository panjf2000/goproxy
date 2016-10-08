package handlers

import (
	_ "bufio"
	"bytes"
	"github.com/Sirupsen/logrus"
	"github.com/panjf2000/goproxy/cache"
	"github.com/panjf2000/goproxy/interface"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

var cachePool api.CachePool
var cacheLog *logrus.Logger

func init() {
	var filename string = "logs/cache.log"
	cacheLog = logrus.New()
	// Log as JSON instead of the default ASCII formatter.
	cacheLog.Formatter = &logrus.TextFormatter{}

	// Output to stderr instead of stdout, could also be a file.
	if cache.CheckFileIsExist(filename) {
		f, err := os.OpenFile(filename, os.O_APPEND, 0666)
		if err != nil {
			return
		}
		cacheLog.Out = f
	} else {
		f, err := os.Create(filename)
		if err != nil {
			return
		}
		cacheLog.Out = f
	}

	// Only log the warning severity or above.
	cacheLog.Level = logrus.DebugLevel

}

func RegisterCachePool(c api.CachePool) {
	cachePool = c
}

//CacheHandler handles "Get" request
func (goproxy *ProxyServer) CacheHandler(rw http.ResponseWriter, req *http.Request) {

	var uri = req.RequestURI

	c := cachePool.Get(uri)

	if c != nil {
		if c.Verify() {
			cacheLog.WithFields(logrus.Fields{
				"request url":    uri,
			}).Debug("Found cache!")
			c.WriteTo(rw)
			return
		} else {
			cacheLog.WithFields(logrus.Fields{
				"request url":    uri,
			}).Debug("Delete cache!")
			cachePool.Delete(uri)
		}
	}

	RmProxyHeaders(req)
	resp, err := goproxy.Travel.RoundTrip(req)
	if err != nil {
		http.Error(rw, err.Error(), 500)
		return
	}
	defer resp.Body.Close()

	cresp := new(http.Response)
	*cresp = *resp
	CopyResponse(cresp, resp)

	cacheLog.WithFields(logrus.Fields{
		"request url":    uri,
	}).Debug("Check out this cache and then stores it if it is right!")
	go cachePool.CheckAndStore(uri, cresp)

	ClearHeaders(rw.Header())
	CopyHeaders(rw.Header(), resp.Header)

	rw.WriteHeader(resp.StatusCode) //写入响应状态

	nr, err := io.Copy(rw, resp.Body)
	if err != nil && err != io.EOF {
		cacheLog.WithFields(logrus.Fields{
			"client": goproxy.Browser,
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
