package handlers

import (
	"net/http"
)

type BaseHandler struct {
	Port string
}

func (p *BaseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, world!"))
}
