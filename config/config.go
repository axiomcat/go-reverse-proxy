package config

import (
	"os"

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
