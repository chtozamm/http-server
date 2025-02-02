package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"html/template"
	"log"
	"net/http"
	"os"
	"slices"
	"time"
)

type application struct {
	auth struct {
		username string
		password string
	}
}

func main() {
	app := new(application)

	app.auth.username = os.Getenv("AUTH_USERNAME")
	if app.auth.username == "" {
		log.Fatal("Missing AUTH_USERNAME environmental variable")
	}

	app.auth.password = os.Getenv("AUTH_PASSWORD")
	if app.auth.password == "" {
		log.Fatal("Missing AUTH_PASSWORD environmental variable")
	}

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("GET /{$}", rootHandler)
	mux.HandleFunc("GET /api/v1/posts", getPosts)
	mux.HandleFunc("GET /api/v1/posts/{id}", getPost)
	mux.HandleFunc("POST /api/v1/posts", app.basicAuth(createPost))
	mux.HandleFunc("PUT /api/v1/posts/{id}", app.basicAuth(updatePost))
	mux.HandleFunc("DELETE /api/v1/posts/{id}", app.basicAuth(deletePost))
	withRequestLogger := requestLogger((mux))

	httpsServer := &http.Server{
		Addr:              ":443",
		Handler:           withRequestLogger,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	go func() {
		log.Println("HTTPS Server is listening on", httpsServer.Addr)
		log.Fatal(httpsServer.ListenAndServeTLS("certs/localhost.pem", "certs/localhost-key.pem"))
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

func (app *application) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte(app.auth.username))
			expectedPasswordHash := sha256.Sum256([]byte(app.auth.password))

			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}
