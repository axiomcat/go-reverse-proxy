package proxy

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/axiomcat/reverse-proxy/config"
	"github.com/axiomcat/reverse-proxy/logger"
)

type ReverseProxy struct {
	tcpProxy         TcpProxy
	httpProxyHandler HttpProxyRequestHandler
	proxyConfig      config.ReverseProxyConfig
	ReloadPort       string
	ConfigPath       string
}

func (r *ReverseProxy) SetupConfig() {
	proxyConfig, err := config.ReadProxyConfig(r.ConfigPath)
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

	go r.StartReloadEndpoint()
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

func (rp *ReverseProxy) StartReloadEndpoint() {
	logger := logger.GetInstance(0)
	mux := http.NewServeMux()
	mux.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		rp.ReloadConfig()
		fmt.Fprintln(w, "Config reloaded")
	})

	httpServer := &http.Server{
		Addr:    rp.ReloadPort,
		Handler: mux,
	}

	go func() {
		logger.Log(fmt.Sprint("Running reload endnpoint on port", rp.ReloadPort))
		httpServer.ListenAndServe()
	}()
}
