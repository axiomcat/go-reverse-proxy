package main

import (
	"github.com/axiomcat/reverse-proxy/logger"
	"github.com/axiomcat/reverse-proxy/proxy"
	"os"
	"os/signal"
)

func main() {
	configPath := "config/config.yml"
	reloadPort := ":42007"

	reverseProxy := proxy.ReverseProxy{ReloadPort: reloadPort, ConfigPath: configPath}

	reverseProxy.SetupConfig()

	go reverseProxy.Start()

	logger := logger.GetInstance(0)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	logger.Log("Recieved interrupt, stopping server")

	reverseProxy.Stop()

	logger.Log("Server shutdown gracefully")
}
