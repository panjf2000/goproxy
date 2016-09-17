package handlers

import (
	"net/http"
)

type BaseHandler struct {
	Protocol string
	Host     []string
}

func (p *BaseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, world!"))
}
