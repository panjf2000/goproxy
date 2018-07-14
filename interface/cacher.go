package api

import (
	"net/http"
	"time"
)

type CachePool interface {
	Get(uri string) Cache
	Delete(uri string)
	CheckAndStore(uri string, req *http.Request, resp *http.Response)
	Clear(d time.Duration)
}

type Cache interface {
	Verify() bool
	WriteTo(rw http.ResponseWriter) (int, error)
}
