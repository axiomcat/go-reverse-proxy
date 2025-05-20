package proxy

import (
	"context"
	"sync"

	"github.com/axiomcat/reverse-proxy/config"
	"github.com/axiomcat/reverse-proxy/logger"
)

type ReverseProxy struct {
	tcpProxy         TcpProxy
	httpProxyHandler HttpProxyRequestHandler
	proxyConfig      config.ReverseProxyConfig
	InternalApiPort  string
	ConfigFile       string
}

func (r *ReverseProxy) SetupConfig() {
	proxyConfig, err := config.ReadProxyConfig(r.ConfigFile)
	r.proxyConfig = proxyConfig
	logger := logger.GetInstance(config.GetLogLevel(proxyConfig))
	logger.UpdateLogLevel(config.GetLogLevel(proxyConfig))

	if err != nil {
		logger.Fatal("Error reading proxy config:", err)
	}
}

func (r *ReverseProxy) ReloadConfig() {
	logger := logger.GetInstance(0)
	logger.Log("Reloading config")

	logger.Log("Shutting down server")

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		r.Stop()
	}()

	wg.Wait()

	logger.Log("Reading config again")

	r.SetupConfig()

	go r.Start()
}

func (r *ReverseProxy) Start() {
	r.tcpProxy = TcpProxy{}

	if r.proxyConfig.Tcp != nil {
		r.tcpProxy = TcpProxy{
			Port:       r.proxyConfig.Tcp.Port,
			TargetAddr: r.proxyConfig.Tcp.Target,
		}

		go r.tcpProxy.Start()
	}

	r.httpProxyHandler = HttpProxyRequestHandler{}

	if r.proxyConfig.HttpRoutes != nil {
		httpProxies := []HttpProxy{}
		for _, httpProxyConfig := range r.proxyConfig.HttpRoutes {
			httpProxy := HttpProxy{
				TargetAddr: httpProxyConfig.Target,
				Host:       httpProxyConfig.Host,
				PrefixPath: httpProxyConfig.PathPrefix,
			}
			httpProxies = append(httpProxies, httpProxy)
		}

		r.httpProxyHandler.HttpProxies = httpProxies
		r.httpProxyHandler.Port = r.proxyConfig.HttpConfig.Port

		go r.httpProxyHandler.Start()
	}
}

func (r *ReverseProxy) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), r.proxyConfig.HttpConfig.ShutdownTimeout)
	defer cancel()

	if r.proxyConfig.HttpRoutes != nil {
		r.httpProxyHandler.Stop(ctx)
	}

	if r.proxyConfig.Tcp != nil {
		r.tcpProxy.Stop()
	}
}

func (r *ReverseProxy) GetNumberOfTcpConnections() int {
	return len(r.tcpProxy.connections)
}
