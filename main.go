package main

import (
	"os"
	"os/signal"

	"github.com/axiomcat/reverse-proxy/api"
	"github.com/axiomcat/reverse-proxy/logger"
	"github.com/axiomcat/reverse-proxy/metrics"
	"github.com/axiomcat/reverse-proxy/proxy"
)

func main() {
	configPath := "config/config.yml"
	auxPort := ":42007"

	reverseProxy := proxy.ReverseProxy{AuxPort: auxPort, ConfigPath: configPath}

	reverseProxy.SetupConfig()

	metrics.CreateInstance()

	go api.StartEndpoints(auxPort, &reverseProxy)

	go reverseProxy.Start()

	logger := logger.GetInstance(0)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, os.Kill)

	<-stop

	logger.Log("Recieved interrupt, stopping server")

	reverseProxy.Stop()

	logger.Log("Server shutdown gracefully")
}
