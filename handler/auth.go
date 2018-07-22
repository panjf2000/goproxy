package handler

import (
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/panjf2000/goproxy/config"
	"github.com/panjf2000/goproxy/tool"
)

var HTTP407 = []byte("HTTP/1.1 407 Proxy Authorization Required\r\nProxy-Authenticate: Basic realm=\"Secure Proxys\"\r\n\r\n")
var authLog *logrus.Logger

func init() {
	logPath := config.RuntimeViper.GetString("server.log_path")
	os.MkdirAll(logPath, os.ModePerm)
	authLog, _ = tool.InitLog(path.Join(logPath, "auth.log"))

}

//Auth provides basic authorization for proxy server.
func (ps *ProxyServer) Auth(rw http.ResponseWriter, req *http.Request) bool {
	var err error
	if config.RuntimeViper.GetBool("server.auth") {
		//代理服务器登入认证
		if ps.Browser, err = ps.auth(rw, req); err != nil {
			authLog.Error("Fail to log in!")
			authLog.WithFields(logrus.Fields{
				"error": err,
			}).Error("Fail to log in!")
			//ps.Browser = "Anonymous"
			return false
		}
		authLog.Info("authentication is passed!")
		return true
	}
	ps.Browser = "Anonymous"
	return true

}

//Auth provides basic authorization for proxy server.
func (ps *ProxyServer) auth(rw http.ResponseWriter, req *http.Request) (string, error) {

	auth := req.Header.Get("Proxy-Authorization")
	auth = strings.Replace(auth, "Basic ", "", 1)

	if auth == "" {
		NeedAuth(rw, HTTP407)
		return "", errors.New("need proxy authorization")
	}
	data, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		authLog.WithFields(logrus.Fields{
			"auth":  auth,
			"error": err,
		}).Error("Fail to decoding Proxy-Authorization!")
		return "", errors.New("fail to decoding Proxy-Authorization")
	}

	var user, passwd string

	userPasswdPair := strings.Split(string(data), ":")
	if len(userPasswdPair) != 2 {
		NeedAuth(rw, HTTP407)
		return "", errors.New("fail to log in")
	}
	user = userPasswdPair[0]
	passwd = userPasswdPair[1]
	if Check(user, passwd) == false {
		NeedAuth(rw, HTTP407)
		return "", errors.New("fail to log in")
	}
	return user, nil
}

func NeedAuth(rw http.ResponseWriter, challenge []byte) error {
	hj, _ := rw.(http.Hijacker)
	Client, _, err := hj.Hijack()
	if err != nil {
		return errors.New("fail to get Tcp connection of client")
	}
	defer Client.Close()

	Client.Write(challenge)
	return nil
}

// Check checks username and password
func Check(user, passwd string) bool {
	if user != "" && passwd != "" {
		if pass, ok := config.RuntimeViper.GetStringMapString("server.user")[user]; ok && pass == passwd {
			return true
		}
	}
	return false
}
