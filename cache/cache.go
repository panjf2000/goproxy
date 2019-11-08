package cache

import (
	// "io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/panjf2000/goproxy/config"
)

type HttpCache struct {
	Header       http.Header `json:"header"`
	Body         []byte      `json:"body"`
	StatusCode   int         `json:"status_code"`
	URI          string      `json:"url"`
	LastModified string      `json:"last_modified"` //eg:"Fri, 27 Jun 2014 07:19:49 GMT"
	ETag         string      `json:"etag"`
	MustVerified bool        `json:"must_verified"`
	//Validity is a time when to verify the cache again.
	Validity time.Time `json:"validity"`
	maxAge   int64     `json:"-"`
}

func NewCacheResp(resp *http.Response) *HttpCache {
	c := new(HttpCache)
	c.Header = make(http.Header)
	CopyHeaders(c.Header, resp.Header)
	c.StatusCode = resp.StatusCode

	var err error
	c.Body, err = ioutil.ReadAll(resp.Body)

	if c.Header == nil {
		return nil
	}

	c.ETag = c.Header.Get("ETag")
	c.LastModified = c.Header.Get("Last-Modified")

	cacheControl := c.Header.Get("Cache-Control")

	// no-cache means you should verify data before use cache.
	// only use cache when remote server returns 302 status.
	if strings.Index(cacheControl, "no-cache") != -1 ||
		strings.Index(cacheControl, "must-revalidate") != -1 ||
		strings.Index(cacheControl, "proxy-revalidate") != -1 {
		c.MustVerified = false
		return nil
	}
	c.MustVerified = true

	if Expires := c.Header.Get("Expires"); Expires != "" {
		c.Validity, err = time.Parse(http.TimeFormat, Expires)
		if err != nil {
			return nil
		}
		log.Println("expire:", c.Validity)
	}

	maxAge := getAge(cacheControl)
	if maxAge != -1 {
		var Time time.Time
		date := c.Header.Get("Date")
		if date == "" {
			Time = time.Now().UTC()
		} else {
			Time, err = time.Parse(time.RFC1123, date)
			if err != nil {
				return nil
			}
		}
		c.Validity = Time.Add(time.Duration(maxAge) * time.Second)
		c.maxAge = maxAge
	} else {
		//c.maxAge, max_age = 0.1 * 60 * 60, 0.1 * 60 * 60
		cacheTimeout := config.RuntimeViper.GetInt64("server.cache_timeout")
		c.maxAge, maxAge = cacheTimeout, cacheTimeout
		Time := time.Now().UTC()
		c.Validity = Time.Add(time.Duration(maxAge) * time.Second)
	}
	log.Println("all:", c.Validity)

	return c
}

// Verify verifies whether cache is out of date.
func (c *HttpCache) Verify() bool {
	if c.MustVerified == true && c.Validity.After(time.Now().UTC()) {
		return true
	}

	newReq, err := http.NewRequest("GET", c.URI, nil)
	if err != nil {
		return false
	}

	if c.LastModified != "" {
		newReq.Header.Add("If-Modified-Since", c.LastModified)
	}
	if c.ETag != "" {
		newReq.Header.Add("If-None-Match", c.ETag)
	}
	Tr := &http.Transport{Proxy: http.ProxyFromEnvironment}
	resp, err := Tr.RoundTrip(newReq)
	if err != nil {
		return false
	}

	if resp.StatusCode != http.StatusNotModified {
		return false
	}
	return false
}

// CacheHandler handles "Get" request
func (c *HttpCache) WriteTo(rw http.ResponseWriter) (int, error) {

	CopyHeaders(rw.Header(), c.Header)
	rw.WriteHeader(c.StatusCode)

	return rw.Write(c.Body)

}

// CopyHeaders copies headers from source to destination.
// Nothing would be returned.
func CopyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

//getAge from Cache Control get cache's lifetime.
func getAge(cacheControl string) (age int64) {
	f := func(sage string) int64 {
		var tmpAge int64
		idx := strings.Index(cacheControl, sage)
		if idx != -1 {
			for i := idx + len(sage) + 1; i < len(cacheControl); i++ {
				if cacheControl[i] >= '0' && cacheControl[i] <= '9' {
					tmpAge = tmpAge*10 + int64(cacheControl[i])
				} else {
					break
				}
			}
			return tmpAge
		}
		return -1
	}
	if sMaxAge := f("s-maxage"); sMaxAge != -1 {
		return sMaxAge
	}
	return f("max-age")
}
