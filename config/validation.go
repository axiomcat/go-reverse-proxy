package config

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"time"
)

type ReverseProxyConfig struct {
	HttpRoutes []*struct {
		Target          string `yaml:"target"`
		Host            string `yaml:"host"`
		PathPrefix      string `yaml:"path_prefix"`
		StripPathPrefix bool   `yaml:"strip_path_prefix"`
	} `yaml:"http_routes"`
	HttpConfig struct {
		Port            string        `yaml:"port"`
		ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	} `yaml:"http_config"`
	LogLevel string `yaml:"log_level"`
}

func validateHTTPURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New(fmt.Sprintf("Invalid HTTP URL, scheme %s is not http nor https", u.Scheme))
	}

	if u.Host == "" {
		return errors.New("Invalid HTTP URL, empty host")
	}

	return nil
}

func ValidateProxyConfig(config ReverseProxyConfig) error {
	for _, httpConfig := range config.HttpRoutes {
		if httpConfig != nil {
			if httpConfig.Target == "" {
				return errors.New("Target url is required for HTTP configuration")
			}
			err := validateHTTPURL(httpConfig.Target)
			if err != nil {
				return err
			}
			if len(httpConfig.PathPrefix) > 0 && httpConfig.PathPrefix[0] != '/' {
				return errors.New("Prefix path is not valid")
			}
		}
	}

	validLogLevels := []string{"debug", "log"}
	if !slices.Contains(validLogLevels, config.LogLevel) {
		return errors.New(fmt.Sprintf("LogLevel %s is not recognized\n", config.LogLevel))
	}

	return nil
}
