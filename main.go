package main

import (
	"github.com/axiomcat/reverse-proxy/proxy"
)

func main() {
	// mux := http.NewServeMux()
	// mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprint(w, "Hello from root!")
	// })
	//
	// httpServer := &http.Server{
	// 	Addr:    ":8021",
	// 	Handler: mux,
	// }
	//
	// go httpServer.ListenAndServe()
	// if err != nil {
	// 	panic("help")
	// }
	tcpProxy := proxy.TcpProxy{
		Port:       ":8020",
		TargetAddr: "localhost:8080",
	}

	go tcpProxy.Start()
	for {

	}
}
