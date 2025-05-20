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
	logger := logger.GetInstance(0)

	var configFile string
	configFile = os.Getenv("CONFIG_FILE")
	if configFile == "" {
		logger.Log("Env var CONFIG_FILE not set, defaulting to config/config.yml")
		configFile = "config/config.yml"
	}

	var internalApiPort string
	internalApiPort = os.Getenv("INTERNAL_API_PORT")
	if internalApiPort == "" {
		logger.Log("Env var INTERNAL_API_PORT not set, defaulting to :42007")
		internalApiPort = ":42007"
	}

	reverseProxy := proxy.ReverseProxy{InternalApiPort: internalApiPort, ConfigFile: configFile}

	reverseProxy.SetupConfig()

	metrics.CreateInstance()

	go api.StartEndpoints(internalApiPort, &reverseProxy)

	go reverseProxy.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, os.Kill)

	<-stop

	logger.Log("Recieved interrupt, stopping server")

	reverseProxy.Stop()

	logger.Log("Server shutdown gracefully")
}
