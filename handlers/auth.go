package handlers

import (
	"encoding/base64"
	"errors"
	"github.com/panjf2000/goproxy/tool"
	"net/http"
	"strings"
)

var HTTP_407 = []byte("HTTP/1.1 407 Proxy Authorization Required\r\nProxy-Authenticate: Basic realm=\"Secure Proxys\"\r\n\r\n")

//Auth provides basic authorizaton for proxy server.
func (goproxy *ProxyServer) Auth(rw http.ResponseWriter, req *http.Request) bool {
	var err error
	if conf.Auth == true {
		//代理服务器登入认证
		if goproxy.Browser, err = goproxy.auth(rw, req); err != nil {
			//log.Debug("%v", err)
			goproxy.Browser = "Anonymous"
			return false
		} else {
			return true
		}
	} else {
		goproxy.Browser = "Anonymous"
		return true
	}

	return false
}

//Auth provides basic authorizaton for proxy server.
func (goproxy *ProxyServer) auth(rw http.ResponseWriter, req *http.Request) (string, error) {

	auth := req.Header.Get("Proxy-Authorization")
	auth = strings.Replace(auth, "Basic ", "", 1)

	if auth == "" {
		NeedAuth(rw, HTTP_407)
		return "", errors.New("Need Proxy Authorization!")
	}
	data, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		//log.Debug("when decoding %v, got an error of %v", auth, err)
		return "", errors.New("Fail to decoding Proxy-Authorization")
	}

	var user, passwd string

	userPasswdPair := strings.Split(string(data), ":")
	if len(userPasswdPair) != 2 {
		NeedAuth(rw, HTTP_407)
		return "", errors.New("Fail to log in")
	} else {
		user = userPasswdPair[0]
		passwd = userPasswdPair[1]
	}
	if Check(user, passwd) == false {
		NeedAuth(rw, HTTP_407)
		return "", errors.New("Fail to log in")
	}
	return user, nil
}

func NeedAuth(rw http.ResponseWriter, challenge []byte) error {
	hj, _ := rw.(http.Hijacker)
	Client, _, err := hj.Hijack()
	if err != nil {
		return errors.New("Fail to get Tcp connection of Client")
	}
	defer Client.Close()

	Client.Write(challenge)
	return nil
}

// Check checks username and password
func Check(user, passwd string) bool {
	if user != "" && passwd != "" && conf.User[user] == passwd {
		return true
	} else {
		return false
	}
}
