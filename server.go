package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/panjf2000/goproxy/handlers"
	_ "net/http"
)

func main() {
	goproxy := handlers.NewProxyServer()

	log.Infof("start proxy server in port:%d", 8080)
	log.Fatal(goproxy.ListenAndServe())
}
