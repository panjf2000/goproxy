package handler

import (
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/lafikl/liblb/bounded"
	"github.com/lafikl/liblb/p2c"
	"github.com/lafikl/liblb/r2"
	"github.com/panjf2000/goproxy/config"
	"github.com/panjf2000/goproxy/tool"
)

var r2LB *r2.R2
var p2cLB *p2c.P2C
var boundedLB *bounded.Bounded
var memcacheServers map[string]int
var serverNodes []string

func init() {
	memcacheServers = make(map[string]int)
	proxyPasses := config.RuntimeViper.GetStringSlice("server.proxy_pass")
	for _, val := range proxyPasses {
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
	r2LB = r2.New(serverNodes...)
	for host, weight := range memcacheServers {
		r2LB.AddWeight(host, weight)
	}
	p2cLB = p2c.New(serverNodes...)
	boundedLB = bounded.New(serverNodes...)
}

func (ps *ProxyServer) Done(req *http.Request) {
	switch config.RuntimeViper.GetInt("server.inverse_mode") {
	case 2:
		p2cLB.Done(req.Host)
	case 3:
		boundedLB.Done(req.Host)
	default:
	}
}

//ReverseHandler handles request for reverse proxy.
//处理反向代理请求
func (ps *ProxyServer) LoadBalancing(req *http.Request) {
	if config.RuntimeViper.GetBool("server.reverse") {
		//用于反向代理，负载均衡
		ps.loadBalancing(req)
	}
}

//ReverseHandler handles request for reverse proxy.
//处理反向代理负载均衡请求
func (ps *ProxyServer) loadBalancing(req *http.Request) {
	var proxyHost string
	mode := config.RuntimeViper.GetInt("server.inverse_mode")
	switch mode {
	case 0:
		// 随机选取一个负载均衡的服务器
		index := tool.GenRandom(0, len(serverNodes), 1)[0]
		proxyHost = serverNodes[index]
	case 1:
		// 轮询法选择反向服务器，支持权重
		proxyHost, _ = r2LB.Balance()
	case 2:
		// power of two choices (p2c)负载均衡算法选择反向服务器
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
		// 边界一致性哈希算法选择反向服务器
		proxyHost, _ = boundedLB.Balance(req.RemoteAddr)
	}
	req.Host = proxyHost
	req.URL.Host = proxyHost
	req.URL.Scheme = "http"
}
