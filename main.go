package main

import (
	"crypto/tls"
	"html/template"
	"log"
	"net/http"
	"slices"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("GET /{$}", rootHandler)
	mux.HandleFunc("GET /api/v1/posts", getPosts)
	mux.HandleFunc("GET /api/v1/posts/{id}", getPost)
	mux.HandleFunc("POST /api/v1/posts", createPost)
	mux.HandleFunc("PUT /api/v1/posts/{id}", updatePost)
	mux.HandleFunc("DELETE /api/v1/posts/{id}", deletePost)
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
			if r.Method != http.MethodGet && r.Method != http.MethodHead {
				http.Error(w, "HTTPS is required for this request method.", http.StatusUpgradeRequired)
				return
			}

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
	mu.Lock()
	defer mu.Unlock()

	var postList []post
	for _, post := range posts {
		postList = append(postList, post)
	}

	slices.SortFunc(postList, func(a, b post) int {
		if a.ID == b.ID {
			return 0
		} else if a.ID < b.ID {
			return 1
		} else {
			return -1
		}
	})

	tmpl, err := template.ParseFiles("static/index.html")
	if err != nil {
		log.Println("Failed to load template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, postList); err != nil {
		log.Println("Failed to render template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
