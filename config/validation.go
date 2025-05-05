package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
)

type ReverseProxyConfig struct {
	Tcp *struct {
		Port   string `yaml:"port"`
		Target string `yaml:"target"`
	} `yaml:"tcp"`
	Http *struct {
		Port   string `yaml:"port"`
		Target string `yaml:"target"`
	} `yaml:"http"`
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

func validateTCPAddress(addr string) error {
	_, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	return nil
}

func validatePort(port string) error {
	if port[0] != ':' {
		return errors.New(fmt.Sprintf("Port definition must start with ':' on %s", port))
	}
	integertPort := port[1:]
	_, err := strconv.Atoi(integertPort)
	if err != nil {
		return err
	}
	return nil
}

func ValidateProxyConfig(config ReverseProxyConfig) error {
	if config.Http != nil {
		if config.Http.Port == "" {
			return errors.New("Port is required for HTTP configuration")
		}
		if config.Http.Target == "" {
			return errors.New("Target url is required for HTTP configuration")
		}
		err := validatePort(config.Http.Port)
		if err != nil {
			return err
		}
		err = validateHTTPURL(config.Http.Target)
		if err != nil {
			return err
		}
	}

	if config.Tcp != nil {
		if config.Tcp.Port == "" {
			return errors.New("Port is required for TCP configuration")
		}
		if config.Tcp.Target == "" {
			return errors.New("Target url is required for TCP configuration")
		}
		err := validatePort(config.Tcp.Port)
		if err != nil {
			return err
		}
		err = validateTCPAddress(config.Tcp.Target)
		if err != nil {
			return err
		}
	}

	if config.Tcp != nil && config.Http != nil {
		if config.Tcp.Port == config.Http.Port {
			return errors.New("TCP port and HTTP port are the same, please change one of them")
		}

	}

	return nil
}
