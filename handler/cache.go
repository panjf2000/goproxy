package handler

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/panjf2000/goproxy/interface"
)

var cachePool api.CachePool

func RegisterCachePool(c api.CachePool) {
	cachePool = c
}

//CacheHandler handles "Get" request
func (ps *ProxyServer) CacheHandler(rw http.ResponseWriter, req *http.Request) {

	uri := req.RequestURI

	c := cachePool.Get(uri)

	if c != nil {
		if c.Verify() {
			_, _ = c.WriteTo(rw)
			return
		}
		cachePool.Delete(uri)
	}

	RmProxyHeaders(req)
	resp, err := ps.Travel.RoundTrip(req)
	if err != nil {
		http.Error(rw, err.Error(), 500)
		return
	}
	defer resp.Body.Close()

	httpResp := new(http.Response)
	*httpResp = *resp
	CopyResponse(httpResp, resp)

	go cachePool.CheckAndStore(uri, req, httpResp)

	ClearHeaders(rw.Header())
	CopyHeaders(rw.Header(), resp.Header)

	rw.WriteHeader(resp.StatusCode) // writes the response status.

	_, err = io.Copy(rw, resp.Body)
	if err != nil && err != io.EOF {
		return
	}
}

func CopyResponse(dest *http.Response, src *http.Response) {
	*dest = *src
	var bodyBytes []byte

	if src.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(src.Body)
	}

	// Restores the io.ReadCloser to its original state
	src.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	dest.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
}
