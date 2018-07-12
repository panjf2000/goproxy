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
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/panjf2000/goproxy/config"
	"github.com/panjf2000/goproxy/handler"
	"github.com/parnurzeal/gorequest"
)

func revRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Fprint(w, "{GET} return: "+r.FormValue("get_req"))
	case "POST":
		fmt.Fprint(w, "{POST} return: "+r.PostFormValue("post_req"))
	default:
		fmt.Fprint(w, "defalut return: nil")
	}
}

func TestServer(t *testing.T) {
	http.HandleFunc("/test_proxy", revRequest)
	for _, addr := range config.RuntimeViper.GetStringSlice("server.proxy_pass") {
		go http.ListenAndServe(addr, nil)
	}

	server := handler.NewProxyServer()
	go server.ListenAndServe()

	resp, body, errs := gorequest.New().Get("http://127.0.0.1/test_proxy").Param("get_req", "Hello World!").End()
	if errs != nil {
		t.Log(errs)
		os.Exit(-1)
	}
	if resp.StatusCode != http.StatusOK {
		t.Logf("response status err, status code:%d\n", resp.StatusCode)
		os.Exit(-1)
	}
	t.Logf("{GET} response: %s\n", body)

	resp, body, errs = gorequest.New().Post("http://127.0.0.1/test_proxy").Send(`{"post_req": "Hello World!"}`).End()

	if errs != nil {
		t.Log(errs)
		os.Exit(-1)
	}
	if resp.StatusCode != http.StatusOK {
		t.Logf("response status err, status code:%d\n", resp.StatusCode)
	}
	t.Logf("{POST} response: %s\n", body)

}
