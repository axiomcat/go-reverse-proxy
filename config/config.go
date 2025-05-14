package config

import (
	"os"

	"github.com/axiomcat/reverse-proxy/logger"
	"gopkg.in/yaml.v3"
)

func ReadProxyConfig(configPath string) (ReverseProxyConfig, error) {
	configData, err := os.ReadFile(configPath)

	if err != nil {
		return ReverseProxyConfig{}, err
	}

	proxyConfig := ReverseProxyConfig{}

	err = yaml.Unmarshal(configData, &proxyConfig)

	if err != nil {
		return ReverseProxyConfig{}, err
	}

	err = ValidateProxyConfig(proxyConfig)
	if err != nil {
		return ReverseProxyConfig{}, err
	}
	return proxyConfig, nil
}

func GetLogLevel(config ReverseProxyConfig) logger.LogLevel {
	if config.LogLevel == "debug" {
		return logger.Debug
	}
	return logger.Logging
}
