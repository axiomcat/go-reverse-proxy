package api

import (
	"fmt"
	"net/http"

	"github.com/axiomcat/reverse-proxy/logger"
	"github.com/axiomcat/reverse-proxy/metrics"
	"github.com/axiomcat/reverse-proxy/proxy"
)

func StartEndpoints(auxPort string, rp *proxy.ReverseProxy) {
	m := metrics.GetInstance()
	logger := logger.GetInstance(0)
	mux := http.NewServeMux()

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := m.GetMetrics()
		fmt.Fprintln(w, metrics)
	})

	mux.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		rp.ReloadConfig()
		fmt.Fprintln(w, "Config reloaded")
	})

	httpServer := &http.Server{
		Addr:    auxPort,
		Handler: mux,
	}

	logger.Log(fmt.Sprint("Running internal API on port", auxPort))
	go httpServer.ListenAndServe()
}
