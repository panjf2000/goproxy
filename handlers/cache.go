package handlers

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"github.com/panjf2000/goproxy/interface"
)

var cachePool api.CachePool

func RegisterCacheHolder(c api.CachePool) {
	cachePool = c
}

//CacheHandler handles "Get" request
func (goproxy *ProxyServer) CacheHandler(rw http.ResponseWriter, req *http.Request) {

	var uri = req.RequestURI

	c := cachePool.Get(uri)

	if c != nil {
		if c.Verify() {
			//log.Debug("Get cache of %s", uri)
			c.WriteTo(rw)
			return
		} else {
			//log.Debug("Delete cache of %s", uri)
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

	//log.Debug("Check and store cache of %s", uri)
	go cachePool.CheckAndStore(uri, cresp)

	ClearHeaders(rw.Header())
	CopyHeaders(rw.Header(), resp.Header)

	rw.WriteHeader(resp.StatusCode) //写入响应状态

	nr, err := io.Copy(rw, resp.Body)
	if err != nil && err != io.EOF {
		//log.Error("%v got an error when copy remote response to client.%v\n", goproxy.Browser, err)
		return
	}
	//log.Info("%v Copied %v bytes from %v.\n", goproxy.Browser, nr, req.URL.Host)
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
