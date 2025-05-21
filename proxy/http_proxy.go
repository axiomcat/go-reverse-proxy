package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/axiomcat/reverse-proxy/logger"
	"github.com/axiomcat/reverse-proxy/metrics"
)

type HttpProxy struct {
	TargetAddr      string
	Host            string
	PrefixPath      string
	StripPathPrefix bool
}

type HttpProxyRequestHandler struct {
	HttpProxies []HttpProxy
	Port        string
	server      *http.Server
}

func (handler *HttpProxyRequestHandler) Start() {
	logger := logger.GetInstance(0)
	metrics := metrics.GetInstance()

	hostToTarget := make(map[string][]HttpProxy)
	for _, proxy := range handler.HttpProxies {
		fullHostPath := proxy.Host + handler.Port
		if proxies, ok := hostToTarget[fullHostPath]; ok {
			proxies = append(proxies, proxy)
			hostToTarget[fullHostPath] = proxies
		} else {
			hostToTarget[fullHostPath] = []HttpProxy{proxy}
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		host := r.Host
		proxies, proxyExist := hostToTarget[host]
		if !proxyExist {
			logger.Log(fmt.Sprintf("There is no proxy matching host %s\n", host))
			return
		}

		path := r.URL.String()
		processedRequest := false
		for _, proxy := range proxies {
			if strings.HasPrefix(path, proxy.PrefixPath) {
				logger.Debug(fmt.Sprintf("Found matching prefix of path %s in proxy %v\n", path, proxy))
				proxy.ForwardRequest(w, r)
				processedRequest = true
				break
			}
		}

		if !processedRequest {
			logger.Log(fmt.Sprintf("Did not find any matching rule for path %s\n", path))
		}
		elapsed := time.Since(start).Milliseconds()
		metrics.RequestTimes = append(metrics.RequestTimes, elapsed)
		metrics.RequestCount += 1
	})

	httpServer := &http.Server{
		Addr:         handler.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	handler.server = httpServer

	go func() {
		logger.Log(fmt.Sprint("Running HTTP reverse proxy on port", handler.Port))
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			logger.Fatal(fmt.Sprintf("Error in ListenAndServe %v\n", err))
		}
	}()
}

func (handler *HttpProxyRequestHandler) Stop(ctx context.Context) {
	logger := logger.GetInstance(0)
	logger.Log("Shutting down HTTP proxy")
	if handler.server != nil {
		if err := handler.server.Shutdown(ctx); err != nil {
			logger.Fatal("Server Shutdown Failed:", err)
		}
	}
}

func (p HttpProxy) ForwardRequest(w http.ResponseWriter, r *http.Request) {
	logger := logger.GetInstance(0)
	ctx := r.Context()
	targetHeaders := make(http.Header)
	for k, v := range r.Header {
		targetHeaders[k] = slices.Clone(v)
	}
	targetHeaders.Set("X-Forwarded-For", r.RemoteAddr)
	targetHeaders.Set("X-Forwarded-Proto", "http")
	targetHeaders.Set("X-Forwarded-Host", r.Host)

	targetUrl, err := url.Parse(p.TargetAddr)
	if err != nil {
		logger.Log(fmt.Sprintf("Error while parsing url %s: %v", p.TargetAddr, err))
		return
	}
	fmt.Println("Url path request", r.URL.Path)
	if p.StripPathPrefix {
		targetUrl.Path = strings.Replace(r.URL.Path, p.PrefixPath, "", 1)
	} else {
		targetUrl.Path = r.URL.Path
	}
	fmt.Println("Url path request", targetUrl.Path)
	targetUrl.RawQuery = r.URL.RawQuery

	targetReq, err := http.NewRequestWithContext(ctx, r.Method, targetUrl.String(), r.Body)
	targetReq.Header = targetHeaders
	logger.Log(targetUrl.String())
	logger.Log(fmt.Sprintf("Headers %v\n", targetReq.Header))

	// client := &http.Client{}
	logger.Debug(fmt.Sprint("Making request to target ", targetReq.URL))

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false}, // if you trust the cert
	}
	// resp, err := client.Do(targetReq)
	resp, err := transport.RoundTrip(targetReq)
	if err != nil {
		logger.Log(fmt.Sprint("Error while sending request to target:", err))
		return
	}

	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)

	maps.Copy[map[string][]string, map[string][]string](w.Header(), resp.Header)

	logger.Log(fmt.Sprintf("%s %s → %s (%d)", r.Method, r.URL.String(), targetUrl.String(), resp.StatusCode))

	io.Copy(w, resp.Body)
}
