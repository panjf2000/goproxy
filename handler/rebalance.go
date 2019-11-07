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
var backendServers map[string]int
var serverNodes []string

func init() {
	backendServers = make(map[string]int)
	proxyPasses := config.RuntimeViper.GetStringSlice("server.proxy_pass")
	for _, addr := range proxyPasses {
		if tool.IsHost(addr) {
			backendServers[addr] = 1
			serverNodes = append(serverNodes, addr)
		} else if tool.IsWeightHost(addr) {
			hostPair := strings.Split(addr, "^")
			host := hostPair[0]
			weight, _ := strconv.Atoi(hostPair[1])
			backendServers[host] = weight
			serverNodes = append(serverNodes, host)
		} else {
		}
	}
	r2LB = r2.New(serverNodes...)
	for host, weight := range backendServers {
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
func (ps *ProxyServer) LoadBalancing(req *http.Request) {
	if config.RuntimeViper.GetBool("server.reverse") {
		ps.loadBalancing(req)
	}
}

//ReverseHandler handles request for reverse proxy.
func (ps *ProxyServer) loadBalancing(req *http.Request) {
	var proxyHost string
	mode := config.RuntimeViper.GetInt("server.inverse_mode")
	switch mode {
	case 0:
		// Selects a back-end server base on randomized algorithm.
		index := tool.GenRandom(0, len(serverNodes), 1)[0]
		proxyHost = serverNodes[index]
	case 1:
		// Selects a back-end server base on polling algorithm which supports weight.
		proxyHost, _ = r2LB.Balance()
	case 2:
		// Selects a back-end server base on power of two choices (p2c) algorithm.
		proxyHost, _ = p2cLB.Balance(req.RemoteAddr)
	case 3:
		// Calculates a HashCode using the client ip and forwards this request to a back-end server base on HashCode with
		// weights in config file.
		ring := tool.NewWithWeights(backendServers)
		if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			server, _ := ring.GetNode(clientIP)
			proxyHost = server
		} else {
			proxyHost = serverNodes[rand.Intn(len(serverNodes))]
		}
	case 4:
		// Selects a back-end server base on Bound Consistent Hashing algorithm.
		proxyHost, _ = boundedLB.Balance(req.RemoteAddr)
	default:
		// Selects a back-end server base on randomized algorithm.
		index := tool.GenRandom(0, len(serverNodes), 1)[0]
		proxyHost = serverNodes[index]

	}
	req.Host = proxyHost
	req.URL.Host = proxyHost
	req.URL.Scheme = "http"
}
