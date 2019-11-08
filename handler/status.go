package handler

var (
	HTTP200 = []byte("HTTP/1.1 200 Connection Established\r\n\r\n")
	HTTP407 = []byte("HTTP/1.1 407 Proxy Authorization Required\r\nProxy-Authenticate: Basic realm=\"Access to internal site\"\r\n\r\n")
)
