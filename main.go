package main

import (
	"bufio"
	"github.com/panjf2000/goproxy/handlers"
	"github.com/panjf2000/goproxy/tool"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func readConf(path string) map[string]string {
	fp, err := os.Open(path)
	defer fp.Close()
	if err != nil {
		log.Fatalln(path, err)
	}
	confMap := make(map[string]string)
	br := bufio.NewReader(fp)
	for {
		line, _, c := br.ReadLine()
		lineString := string(line[:])
		if c == io.EOF {
			break
		}
		if 0 == len(lineString) || lineString == "\r\n" || strings.HasPrefix(lineString, "#") {
			continue
		}
		confKey := strings.Split(lineString, "=")[0]
		confKey = strings.TrimSpace(confKey)
		confValue := strings.Split(lineString, "=")[1]
		confValue = strings.TrimSpace(confValue)
		confMap[confKey] = confValue
	}

	return confMap
}

func main() {
	var proxyHandler handlers.ProxyHandler

	// 读取配置文件
	var confMap map[string]string = readConf("conf/proxy.conf")
	if val, ok := confMap["default_protocol"]; ok {
		proxyHandler.Protocol = val
	} else {
		proxyHandler.Protocol = "https://"
	}

	// 利用正则表达式提取出配置文件中的待转发服务器，目前支持ip和域名
	if val, ok := confMap["host_list"]; ok {
		var rx *regexp.Regexp = tool.DomainOrIP
		proxyHandler.Host = rx.FindAllString(val, -1)

	} else {
		proxyHandler.Host = []string{"127.0.0.1:80"}
	}

	// 选择负载策略，目前支持随机和IP Hash两种模式
	if val, ok := confMap["mode"]; ok {
		switch val {
		case "hash":
			proxyHandler.Mode = 0
		case "random":
			proxyHandler.Mode = 1

		}
	} else {
		proxyHandler.Mode = 0
	}
	var listen string
	if val, ok := confMap["listen"]; ok {
		listen = ":" + val
	} else {
		listen = ":8080"
	}

	// 启动http server，监听预设的端口,并在后台进行转发
	http.Handle("/", &proxyHandler)
	if err := http.ListenAndServe(listen, &proxyHandler); err != nil {
		log.Fatalln("ListenAndServe occur a error: ", err)
	}
	select {} //利用select关键字的特性，阻塞主进程，使其成为守护进程
}
