package cache

import (
	"log"
	"net/http"
	"strings"
)

// IsReqCache checks whether request ask to be stored as cache
func IsReqCache(req *http.Request) bool {
	log.Printf("http request header:%v", req.Header)
	cacheControl := req.Header.Get("Cache-Control")
	contentType := req.Header.Get("Content-Type")
	if cacheControl == "" && contentType == "" {
		return true
	} else if len(cacheControl) > 0 {
		if strings.Index(cacheControl, "private") != -1 ||
			strings.Index(cacheControl, "no-cache") != -1 ||
			strings.Index(cacheControl, "no-store") != -1 ||
			strings.Index(cacheControl, "must-revalidate") != -1 ||
			(strings.Index(cacheControl, "max-age") == -1 &&
				strings.Index(cacheControl, "s-maxage") == -1 &&
				req.Header.Get("Etag") == "" &&
				req.Header.Get("Last-Modified") == "" &&
				(req.Header.Get("Expires") == "" || req.Header.Get("Expires") == "0")) {
			return false
		}

	} else if len(contentType) > 0 {
		if strings.Index(contentType, "video") != -1 ||
			strings.Index(contentType, "image") != -1 ||
			strings.Index(contentType, "audio") != -1 {
			return false
		}

	}
	return true
}

// IsRespCache checks whether response can be stored as cache
func IsRespCache(resp *http.Response) bool {
	log.Printf("http response header:%v", resp.Header)
	cacheControl := resp.Header.Get("Cache-Control")
	contentType := resp.Header.Get("Content-Type")
	if cacheControl == "" && contentType == "" {
		return true
	} else if len(cacheControl) > 0 {
		if strings.Index(cacheControl, "private") != -1 ||
			strings.Index(cacheControl, "no-cache") != -1 ||
			strings.Index(cacheControl, "no-store") != -1 ||
			strings.Index(cacheControl, "must-revalidate") != -1 ||
			(strings.Index(cacheControl, "max-age") == -1 &&
				strings.Index(cacheControl, "s-maxage") == -1 &&
				resp.Header.Get("Etag") == "" &&
				resp.Header.Get("Last-Modified") == "" &&
				(resp.Header.Get("Expires") == "" || resp.Header.Get("Expires") == "0")) {
			return false
		}

	} else if len(contentType) > 0 {
		if strings.Index(contentType, "video") != -1 ||
			strings.Index(contentType, "image") != -1 ||
			strings.Index(contentType, "audio") != -1 {
			return false
		}

	}
	return true
}
