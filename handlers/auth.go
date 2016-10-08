package handlers

import (
	"encoding/base64"
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/panjf2000/goproxy/cache"
	_ "github.com/panjf2000/goproxy/tool"
	"net/http"
	"os"
	"strings"
)

var HTTP_407 = []byte("HTTP/1.1 407 Proxy Authorization Required\r\nProxy-Authenticate: Basic realm=\"Secure Proxys\"\r\n\r\n")
var authLog *logrus.Logger

func init() {
	var filename string = "logs/auth.log"
	authLog = logrus.New()
	// Log as JSON instead of the default ASCII formatter.
	authLog.Formatter = &logrus.TextFormatter{}

	// Output to stderr instead of stdout, could also be a file.
	if cache.CheckFileIsExist(filename) {
		f, err := os.OpenFile(filename, os.O_APPEND, 0666)
		if err != nil {
			return
		}
		authLog.Out = f
	} else {
		f, err := os.Create(filename)
		if err != nil {
			return
		}
		authLog.Out = f
	}

	// Only log the warning severity or above.
	authLog.Level = logrus.DebugLevel

}

//Auth provides basic authorizaton for proxy server.
func (goproxy *ProxyServer) Auth(rw http.ResponseWriter, req *http.Request) bool {
	var err error
	if conf.Auth == true {
		//代理服务器登入认证
		if goproxy.Browser, err = goproxy.auth(rw, req); err != nil {
			authLog.Error("Fail to log in!")
			//log.Debug("%v", err)
			//goproxy.Browser = "Anonymous"
			return false
		} else {
			authLog.Info("authentication is passed!")
			return true
		}
	} else {
		goproxy.Browser = "Anonymous"
		return true
	}

	return true
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
		authLog.WithFields(logrus.Fields{
			"auth":  auth,
			"error": err,
		}).Error("Fail to decoding Proxy-Authorization!")
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
