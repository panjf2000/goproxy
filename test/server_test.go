/*
@version: 1.0
@author: allanpan
@license:  Apache Licence
@contact: panjf2000@gmail.com
@site:
@file: server_test.go
@time: 2018/7/4 15:01
@tag: 1,2,3
@todo: ...
*/
package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/panjf2000/goproxy/config"
	"github.com/panjf2000/goproxy/handler"
	"github.com/parnurzeal/gorequest"
)

func init() {
	http.HandleFunc("/test_proxy", revRequest)
	for _, addr := range config.RuntimeViper.GetStringSlice("server.proxy_pass") {
		go func() {
			log.Fatalln(http.ListenAndServe(addr, nil))
		}()
	}

	server := handler.NewProxyServer()
	go func() {
		log.Fatalln(server.ListenAndServe())
	}()
	time.Sleep(5 * time.Second)
}

func revRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Fprint(w, "{GET} return: "+r.FormValue("get_req"))
	case "POST":
		b, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		var m map[string]interface{}
		json.Unmarshal(b, &m)
		postParam, _ := m["post_req"].(string)
		fmt.Fprint(w, "{POST} return: "+postParam)
	default:
		fmt.Fprint(w, "defalut return: nil")
	}
}

func TestServer(t *testing.T) {
	for range []int{1, 2, 3, 4, 5} {
		resp, body, errs := gorequest.New().Get("http://127.0.0.1:8080/test_proxy").Param("get_req", "Hello World!").End()
		if errs != nil {
			t.Fatal(errs)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("response status err, status code:%d\n", resp.StatusCode)
		}
		t.Logf("{GET} response: %s\n", body)
	}
	
	resp, body, errs := gorequest.New().Post("http://127.0.0.1:8080/test_proxy").Send(`{"post_req": "Hello World!"}`).End()

	if errs != nil {
		t.Fatal(errs)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("response status err, status code:%d\n", resp.StatusCode)
	}
	t.Logf("{POST} response: %s\n", body)

}
