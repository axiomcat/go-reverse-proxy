package main

import (
	"log"

	"github.com/axiomcat/reverse-proxy/config"
	"github.com/axiomcat/reverse-proxy/proxy"
)

func main() {
	configPath := "config/config.yml"
	proxyConfig, err := config.ReadProxyConfig(configPath)
	if err != nil {
		log.Fatalln("Error reading proxy config:", err)
	}

	if proxyConfig.Tcp != nil {
		tcpProxy := proxy.TcpProxy{
			Port:       proxyConfig.Tcp.Port,
			TargetAddr: proxyConfig.Tcp.Target,
		}

		go tcpProxy.Start()
	}

	if proxyConfig.Http != nil {
		httpProxies := []proxy.HttpProxy{}
		for _, httpProxyConfig := range proxyConfig.Http {
			httpProxy := proxy.HttpProxy{
				TargetAddr: httpProxyConfig.Target,
				Host:       httpProxyConfig.Host,
			}
			httpProxies = append(httpProxies, httpProxy)
		}

		httpProxyHandler := proxy.HttpProxyRequestHandler{
			HttpProxies: httpProxies,
			Port:        proxyConfig.HttpPort,
		}

		go httpProxyHandler.Start()
	}

	for {

	}
}
