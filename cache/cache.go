//Package cache handlers http web cache.
package cache

import (
    // "io"
    "io/ioutil"
    "log"
    "net/http"
    "strings"
    "time"
)

type Cache struct {
    Header        http.Header `json:"header"`
    Body          []byte      `json:"body"`
    StatusCode    int         `json:"status_code"`
    URI           string      `json:"url"`
    Last_Modified string      `json:"last_modified"` //eg:"Fri, 27 Jun 2014 07:19:49 GMT"
    ETag          string      `json:"etag"`
    Mustverified  bool        `json:"must_verified"`
    //Vlidity is a time when to verfiy the cache again.
    Vlidity time.Time `json:"vlidity"`
    maxAge  int64     `json:"-"`
}

func New(resp *http.Response) *Cache {
    c := new(Cache)
    c.Header = make(http.Header)
    CopyHeaders(c.Header, resp.Header)
    c.StatusCode = resp.StatusCode

    var err error
    c.Body, err = ioutil.ReadAll(resp.Body)

    if c.Header == nil {
        return nil
    }

    c.ETag = c.Header.Get("ETag")
    c.Last_Modified = c.Header.Get("Last-Modified")

    Cache_Control := c.Header.Get("Cache-Control")

    // no-cache means you should verify data before use cache.
    // only use cache when remote server returns 302 status.
    if strings.Index(Cache_Control, "no-cache") != -1 ||
        strings.Index(Cache_Control, "must-revalidate") != -1 ||
        strings.Index(Cache_Control, "proxy-revalidate") != -1 {
        c.Mustverified = true
        return nil
    }

    if Expires := c.Header.Get("Expires"); Expires != "" {
        c.Vlidity, err = time.Parse(http.TimeFormat, Expires)
        if err != nil {
            return nil
        }
        log.Println("expire:", c.Vlidity)
    }

    max_age := getAge(Cache_Control)
    if max_age != -1 {
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
        c.Vlidity = Time.Add(time.Duration(max_age) * time.Second)
        c.maxAge = max_age
    } else {
        c.maxAge = 24 * 60 * 60
    }

    log.Println("all:", c.Vlidity)

    return c
}

// Verify verifies whether cache is out of date.
func (c *Cache) Verify() bool {
    if c.Mustverified == false && c.Vlidity.After(time.Now().UTC()) {
        return true
    }

    newReq, err := http.NewRequest("GET", c.URI, nil)
    if err != nil {
        return false
    }

    if c.Last_Modified != "" {
        newReq.Header.Add("If-Modified-Since", c.Last_Modified)
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
    return true
}

// CacheHandler handles "Get" request
func (c *Cache) WriteTo(rw http.ResponseWriter) (int, error) {

    CopyHeaders(rw.Header(), c.Header)
    rw.WriteHeader(c.StatusCode)

    return rw.Write(c.Body)

}

// CopyHeaders copy headers from source to destination.
// Nothing would be returned.
func CopyHeaders(dst, src http.Header) {
    for key, values := range src {
        for _, value := range values {
            dst.Add(key, value)
        }
    }
}

//getAge from Cache Control get cache's lifetime.
func getAge(Cache_Control string) (age int64) {
    f := func(sage string) int64 {
        var tmpAge int64
        idx := strings.Index(Cache_Control, sage)
        if idx != -1 {
            for i := idx + len(sage) + 1; i < len(Cache_Control); i++ {
                if Cache_Control[i] >= '0' && Cache_Control[i] <= '9' {
                    tmpAge = tmpAge*10 + int64(Cache_Control[i])
                } else {
                    break
                }
            }
            return tmpAge
        }
        return -1
    }
    if s_maxage := f("s-maxage"); s_maxage != -1 {
        return s_maxage
    }
    return f("max-age")
}
