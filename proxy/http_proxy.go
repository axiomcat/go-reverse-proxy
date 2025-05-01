package proxy

import (
	"fmt"
	"net/http"
)

type HttpProxy struct {
	Port       string
	TargetAddr string
}

func (p HttpProxy) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello from root!")
	})

	httpServer := &http.Server{
		Addr:    ":8021",
		Handler: mux,
	}

	go httpServer.ListenAndServe()
}
