package main

import (
	"github.com/panjf2000/goproxy/handlers"
	"log"
	"net/http"
)

func main() {
	var proxyHandler handlers.ProxyHandler
	proxyHandler.Port = "80"
	http.Handle("/", &proxyHandler)
	err := http.ListenAndServe(":8888", &proxyHandler)
	select {}
	if err != nil {
		log.Fatalln("ListenAndServe occur a error: ", err)
	}
}
