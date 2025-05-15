package proxy

import (
	"context"
	"fmt"
	"net/http"

	"github.com/axiomcat/reverse-proxy/config"
	"github.com/axiomcat/reverse-proxy/logger"
)

type ReverseProxy struct {
	TcpProxy         TcpProxy
	HttpProxyHandler HttpProxyRequestHandler
	ProxyConfig      config.ReverseProxyConfig
	ReloadPort       string
	ConfigPath       string
}

func (r *ReverseProxy) SetupConfig() {
	proxyConfig, err := config.ReadProxyConfig(r.ConfigPath)
	r.ProxyConfig = proxyConfig
	logger := logger.GetInstance(config.GetLogLevel(proxyConfig))
	logger.UpdateLogLevel(config.GetLogLevel(proxyConfig))

	if err != nil {
		logger.Fatal("Error reading proxy config:", err)
	}
}

func (r *ReverseProxy) ReloadConfig() {
	logger := logger.GetInstance(0)
	logger.Log("Reloading config")
	ctx, cancel := context.WithTimeout(context.Background(), r.ProxyConfig.HttpConfig.ShutdownTimeout)
	defer cancel()
	logger.Log("Shutting down server")
	if r.HttpProxyHandler.Server != nil {
		r.HttpProxyHandler.Stop(ctx)
	}
	r.SetupConfig()
	go r.Start()
}

func (r *ReverseProxy) Start() {
	r.TcpProxy = TcpProxy{}

	if r.ProxyConfig.Tcp != nil {
		r.TcpProxy = TcpProxy{
			Port:       r.ProxyConfig.Tcp.Port,
			TargetAddr: r.ProxyConfig.Tcp.Target,
		}

		// go r.TcpProxy.Start()
	}

	r.HttpProxyHandler = HttpProxyRequestHandler{}

	if r.ProxyConfig.HttpRoutes != nil {
		httpProxies := []HttpProxy{}
		for _, httpProxyConfig := range r.ProxyConfig.HttpRoutes {
			httpProxy := HttpProxy{
				TargetAddr: httpProxyConfig.Target,
				Host:       httpProxyConfig.Host,
				PrefixPath: httpProxyConfig.PathPrefix,
			}
			httpProxies = append(httpProxies, httpProxy)
		}

		r.HttpProxyHandler.HttpProxies = httpProxies
		r.HttpProxyHandler.Port = r.ProxyConfig.HttpConfig.Port

		go r.HttpProxyHandler.Start()
	}

	go r.StartReloadEndpoint()
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
