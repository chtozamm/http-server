package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", rootHandler)
	withMiddleware := requestLogger((mux))

	cert, err := tls.LoadX509KeyPair("certs/localhost.pem", "certs/localhost-key.pem")
	if err != nil {
		log.Fatalf("Failed to get certificates: %v\n", err)
	}
	TLSConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	httpsServer := &http.Server{
		Addr:              ":443",
		Handler:           withMiddleware,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		TLSConfig:         TLSConfig,
	}

	go func() {
		log.Println("HTTPS Server is listening on", httpsServer.Addr)
		log.Fatal(httpsServer.ListenAndServeTLS("", ""))
	}()

	httpServer := &http.Server{
		Addr:    ":80",
		Handler: httpsRedirectMiddleware(http.NotFoundHandler()),
	}

	log.Println("HTTP Server is listening on :80 for redirection")
	log.Fatal(httpServer.ListenAndServe())
}

func httpsRedirectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Scheme != "https" {
			httpsURL := "https://" + r.Host + r.URL.Path
			if r.URL.RawQuery != "" {
				httpsURL += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Println(rec.statusCode, r.Method, r.URL.Path, r.RemoteAddr)
	})
}

// responseRecorder is a custom ResponseWriter that captures the status code
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *responseRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello there!"))
}
