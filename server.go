package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/panjf2000/goproxy/handlers"
	_ "net/http"
)

func main() {
	goproxy := handlers.NewProxyServer()

	log.Println("start my proxy server!")
	log.Fatal(goproxy.ListenAndServe())
}
