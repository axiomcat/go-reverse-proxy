package proxy

import (
	"io"
	"log"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"time"
)

type HttpProxy struct {
	TargetAddr string
	Host       string
}

type HttpProxyRequestHandler struct {
	HttpProxies []HttpProxy
	Port        string
}

func (handler HttpProxyRequestHandler) Start() {
	hostToTarget := make(map[string]HttpProxy)
	for _, proxy := range handler.HttpProxies {
		fullHostPath := proxy.Host + handler.Port
		hostToTarget[fullHostPath] = proxy
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		proxy, proxyExist := hostToTarget[host]
		if !proxyExist {
			log.Printf("There is no proxy matching host %s\n", host)
			return
		}
		proxy.ForwardRequest(w, r)
	})

	httpServer := &http.Server{
		Addr:         handler.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Println("Running HTTP reverse proxy on port", handler.Port)

	go httpServer.ListenAndServe()
}

func (p HttpProxy) ForwardRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	targetHeaders := make(map[string][]string)
	for k, v := range r.Header {
		targetHeaders[k] = slices.Clone(v)
	}
	forwardedForHeader := targetHeaders["X-Forwarded-For"]
	targetHeaders["X-Forwarded-For"] = append(forwardedForHeader, r.RemoteAddr)
	targetHeaders["X-Forwarded-Proto"] = []string{"http"}
	targetHeaders["X-Forwarded-Host"] = []string{r.Host}

	targetUrl, err := url.Parse(p.TargetAddr)
	if err != nil {
		log.Printf("Error while parsing url %s: %v", p.TargetAddr, err)
		return
	}
	targetUrl.Path = r.URL.Path
	targetUrl.RawQuery = r.URL.RawQuery

	targetReq, err := http.NewRequestWithContext(ctx, r.Method, targetUrl.String(), r.Body)
	targetReq.Header = targetHeaders
	targetReq.Host = r.Host

	client := &http.Client{}
	resp, err := client.Do(targetReq)
	if err != nil {
		log.Println("Error while sending request to target:", err)
		return
	}

	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)

	maps.Copy[map[string][]string, map[string][]string](w.Header(), resp.Header)

	log.Printf("%s %s → %s (%d)", r.Method, r.URL.String(), targetUrl.String(), resp.StatusCode)

	io.Copy(w, resp.Body)
}
