package cache

import (
	"net/http"
	"os"
	"strings"
)

//IsCache checks whether response can be stored as cache
func IsCache(resp *http.Response) bool {

	Cache_Control := resp.Header.Get("Cache-Control")
	Content_type := resp.Header.Get("Content-Type")
	if strings.Index(Cache_Control, "private") != -1 ||
		strings.Index(Cache_Control, "no-store") != -1 ||
		strings.Index(Content_type, "application") != -1 ||
		strings.Index(Content_type, "video") != -1 ||
		strings.Index(Content_type, "audio") != -1 ||
		(strings.Index(Cache_Control, "max-age") == -1 &&
			strings.Index(Cache_Control, "s-maxage") == -1 &&
			resp.Header.Get("Etag") == "" &&
			resp.Header.Get("Last-Modified") == "" &&
			(resp.Header.Get("Expires") == "" || resp.Header.Get("Expires") == "0")) {
		return false
	}
	return true
}
func CheckFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
