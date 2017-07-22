package handlers

import (
	"github.com/panjf2000/goproxy/tool"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"github.com/lafikl/liblb/p2c"
	"github.com/lafikl/liblb/r2"
	"github.com/lafikl/liblb/bounded"
)

var r2LB *r2.R2
var p2cLB *p2c.P2C
var boundedLB *bounded.Bounded
//ReverseHandler handles request for reverse proxy.
//处理反向代理请求
func (ps *ProxyServer) LoadBalancing(req *http.Request) {
	if conf.Reverse == true {
		//用于反向代理
		ps.loadBalancing(req)
	}
}

func (ps *ProxyServer) Done(req *http.Request) {
	switch conf.Mode {
	case 2:
		p2cLB.Done(req.Host)
	case 3:
		boundedLB.Done(req.Host)
	default:
	}
}

//ReverseHandler handles request for reverse proxy.
//处理反向代理请求
func (ps *ProxyServer) loadBalancing(req *http.Request) {
	var proxyHost string
	memcacheServers := make(map[string]int)
	var serverNodes []string
	for _, val := range conf.ProxyPass {
		if tool.IsHost(val) {
			memcacheServers[val] = 1
			serverNodes = append(serverNodes, val)
		} else if tool.IsWeightHost(val) {
			hostPair := strings.Split(val, "^")
			host := hostPair[0]
			weight, _ := strconv.Atoi(hostPair[1])
			memcacheServers[host] = weight
			serverNodes = append(serverNodes, host)
		} else {

		}
	}
	switch conf.Mode {
	case 0:
		// 随机选取一个负载均衡的服务器
		index := rand.Intn(len(serverNodes))
		proxyHost = serverNodes[index]
	case 1:
		if r2LB == nil {
			r2LB = r2.New(serverNodes...)
		}
		for host, weight := range memcacheServers {
			r2LB.AddWeight(host, weight)
		}
		proxyHost, _ = r2LB.Balance()
	case 2:
		if p2cLB == nil {
			p2cLB = p2c.New(serverNodes...)
		}
		proxyHost, _ = p2cLB.Balance(req.RemoteAddr)
	case 3:
		// 根据客户端的IP算出一个HASH值，将请求分配到集群中的某一台服务器上, 依据配置文件中设置的每个服务器的权重进行负载均衡
		ring := tool.NewWithWeights(memcacheServers)
		if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			server, _ := ring.GetNode(clientIP)
			proxyHost = server
		} else {
			proxyHost = serverNodes[rand.Intn(len(serverNodes))]
		}
	case 4:
		if boundedLB == nil {
			boundedLB = bounded.New(serverNodes...)
		}
		proxyHost, _ = boundedLB.Balance(req.RemoteAddr)
	}
	req.Host = proxyHost
	req.URL.Host = proxyHost
	req.URL.Scheme = "http"
}
