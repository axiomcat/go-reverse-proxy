package main

import (
	"github.com/axiomcat/reverse-proxy/proxy"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type ReverseProxyConfig struct {
	Tcp struct {
		Port   string
		Target string
	}
	Http struct {
		Port   string
		Target string
	}
}

func main() {
	configPath := "config/config.yml"
	configData, err := os.ReadFile(configPath)

	if err != nil {
		log.Fatalf("Error reading config from %s: %v\n", configPath, err)
	}
	proxyConfig := ReverseProxyConfig{}

	err = yaml.Unmarshal(configData, &proxyConfig)

	if err != nil {
		log.Fatalf("Error parsing yaml config %v\n", err)
	}

	tcpProxy := proxy.TcpProxy{
		Port:       proxyConfig.Tcp.Port,
		TargetAddr: proxyConfig.Tcp.Target,
	}

	httpProxy := proxy.HttpProxy{
		Port:       proxyConfig.Http.Port,
		TargetAddr: proxyConfig.Http.Target,
	}

	go tcpProxy.Start()
	go httpProxy.Start()

	for {

	}
}
