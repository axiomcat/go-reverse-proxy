package main

import (
	"github.com/axiomcat/reverse-proxy/proxy"
)

func main() {
	tcpProxy := proxy.TcpProxy{
		Port:       ":8020",
		TargetAddr: "localhost:8080",
	}

	httpProxy := proxy.HttpProxy{
		Port:       ":8021",
		TargetAddr: "http://localhost:8081",
	}

	go tcpProxy.Start()
	go httpProxy.Start()

	for {

	}
}
