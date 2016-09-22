package main

import (
	"github.com/panjf2000/goproxy/handlers"
	log "github.com/Sirupsen/logrus"
	_ "net/http"
)

func main() {
	goproxy := handlers.NewProxyServer()

	log.Println("start my proxy server!")
	log.Fatal(goproxy.ListenAndServe())
}
