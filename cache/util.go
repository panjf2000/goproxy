package cache

import (
    "net/http"
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

// //CheckCaches evey certian minutes check whether cache is out of date, if yes release it.
// func CheckCaches() {
//     for {
//         time.Sleep(time.Duration(cnfg.CacheTimeout) * time.Minute)
//         for key, Cache := range Caches {
//             if Cache != nil && Cache.Verify() == false {
//                 Caches.DeleteByCheckSum(key)
//             }
//         }
//     }
// }
