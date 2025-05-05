package proxy

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"slices"
)

type HttpProxy struct {
	Port       string
	TargetAddr string
}

func (p HttpProxy) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
		// body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Error reading response body:", err)
		}
		w.WriteHeader(resp.StatusCode)
		for k, v := range resp.Header {
			w.Header().Set(k, v[0])
		}
		// fmt.Fprint(w, string(body))
		io.Copy(w, resp.Body)
	})

	httpServer := &http.Server{
		Addr:    p.Port,
		Handler: mux,
	}

	go httpServer.ListenAndServe()
}
