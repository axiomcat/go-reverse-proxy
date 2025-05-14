package main

import (
	"github.com/axiomcat/reverse-proxy/config"
	"github.com/axiomcat/reverse-proxy/logger"
	"github.com/axiomcat/reverse-proxy/proxy"
)

func main() {
	configPath := "config/config.yml"
	proxyConfig, err := config.ReadProxyConfig(configPath)

	logger := logger.GetInstance(config.GetLogLevel(proxyConfig))

	if err != nil {
		logger.Fatal("Error reading proxy config:", err)
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
				PrefixPath: httpProxyConfig.PathPrefix,
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
