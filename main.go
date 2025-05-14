package main

import (
	"context"
	"os"
	"os/signal"

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

	tcpProxy := proxy.TcpProxy{}

	if proxyConfig.Tcp != nil {
		tcpProxy = proxy.TcpProxy{
			Port:       proxyConfig.Tcp.Port,
			TargetAddr: proxyConfig.Tcp.Target,
		}

		go tcpProxy.Start()
	}

	httpProxyHandler := proxy.HttpProxyRequestHandler{}

	if proxyConfig.HttpRoutes != nil {
		httpProxies := []proxy.HttpProxy{}
		for _, httpProxyConfig := range proxyConfig.HttpRoutes {
			httpProxy := proxy.HttpProxy{
				TargetAddr: httpProxyConfig.Target,
				Host:       httpProxyConfig.Host,
				PrefixPath: httpProxyConfig.PathPrefix,
			}
			httpProxies = append(httpProxies, httpProxy)
		}

		httpProxyHandler.HttpProxies = httpProxies
		httpProxyHandler.Port = proxyConfig.HttpConfig.Port

		go httpProxyHandler.Start()
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	logger.Log("Recieved interrupt, stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), proxyConfig.HttpConfig.ShutdownTimeout)
	defer cancel()

	if httpProxyHandler.Server != nil {
		httpProxyHandler.Stop(ctx)
	}

	logger.Log("Server shutdown gracefully")
}
