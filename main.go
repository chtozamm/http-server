package main

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	// pb "github.com/chtozamm/http-server/grpc"
)

type application struct {
	auth struct {
		username string
		password string
	}
	db             *pgxpool.Pool
	enabledModules map[string]bool
	// pb.UnimplementedHttpServerServiceServer
}

func main() {
	var err error

	app := new(application)

	app.enabledModules = map[string]bool{}

	cfg, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration from file: %v\n", err)
	}

	for _, module := range cfg.Modules {
		app.enabledModules[module] = true
	}

	if app.enabledModules["database"] {
		// Create concurrency safe database connection pool
		dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Fatalf("Unable to connect to database: %v\n", err)
		}
		defer dbpool.Close()
		app.db = dbpool
	}

	if app.enabledModules["auth"] {
		// Get credentials for basic authentication
		app.auth.username = os.Getenv("AUTH_USERNAME")
		if app.auth.username == "" {
			log.Fatal("Missing AUTH_USERNAME environmental variable")
		}

		app.auth.password = os.Getenv("AUTH_PASSWORD")
		if app.auth.password == "" {
			log.Fatal("Missing AUTH_PASSWORD environmental variable")
		}
	}

	// if app.enabledModules["grpc"] {
	// 	// Start gRPC server
	// 	go func() {
	// 		lis, err := net.Listen("tcp", ":50051")
	// 		if err != nil {
	// 			log.Fatalf("Failed to announce listener: %v", err)
	// 		}
	// 		s := grpc.NewServer()
	// 		pb.RegisterHttpServerServiceServer(s, app)
	// 		if err := s.Serve(lis); err != nil {
	// 			log.Fatalf("Failed to start gRPC server: %v", err)
	// 		}
	// 	}()
	// }

	// Define routes
	mux := http.NewServeMux()
	if app.enabledModules["webui"] {
		mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
		mux.HandleFunc("GET /{$}", app.rootHandler)
	} else {
		mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/api/v1/posts", http.StatusMovedPermanently)
		})
	}
	mux.HandleFunc("GET /api/v1/posts", app.getPosts)
	mux.HandleFunc("GET /api/v1/posts/{id}", app.getPost)
	mux.Handle("POST /api/v1/posts", app.basicAuthMiddleware(enforceJSONMiddleware(app.createPost)))
	mux.Handle("PUT /api/v1/posts/{id}", app.basicAuthMiddleware(enforceJSONMiddleware(app.updatePost)))
	mux.Handle("DELETE /api/v1/posts/{id}", app.basicAuthMiddleware(app.deletePost))
	mux.HandleFunc("GET /api/v1/healthz", app.healthCheckHandler)

	// // Main HTTPS server
	// httpsServer := &http.Server{
	// 	Addr:              ":443",
	// 	Handler:           requestLoggerMiddleware(mux),
	// 	ReadTimeout:       5 * time.Second,
	// 	WriteTimeout:      10 * time.Second,
	// 	IdleTimeout:       30 * time.Second,
	// 	ReadHeaderTimeout: 2 * time.Second,
	// }

	// // Start HTTPS server
	// go func() {
	// 	log.Printf("HTTPS Server is listening on %s", httpsServer.Addr)
	// 	err := httpsServer.ListenAndServeTLS("certs/localhost.pem", "certs/localhost-key.pem")
	// 	if !errors.Is(err, http.ErrServerClosed) {
	// 		log.Fatal(err)
	// 	}
	// }()

	// HTTP server for redirects to HTTPS
	httpServer := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.Port),
		// Handler: requestLoggerMiddleware(httpsRedirectMiddleware(http.NotFoundHandler())),
		Handler:           requestLoggerMiddleware(mux),
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	// Start HTTP server
	go func() {
		// log.Printf("HTTP Server is listening on %s for redirection\n", addr)
		log.Printf("HTTP Server is listening on :%d\n", cfg.Port)
		err := httpServer.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	// Graceful shutdown
	shutdownError := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		log.Printf("Shutting down server: %v\n", s)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// shutdownError <- httpsServer.Shutdown(ctx)
		shutdownError <- httpServer.Shutdown(ctx)
	}()

	err = <-shutdownError
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Server has been stopped")
}

func requestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)
		clientAddr := r.Header.Get("X-Client-IP")
		log.Println(rec.statusCode, r.Method, r.URL.Path, r.RemoteAddr, clientAddr)
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

func (app *application) rootHandler(w http.ResponseWriter, r *http.Request) {
	postList, err := app.getPostsList()
	if err != nil {
		log.Printf("Failed to get posts: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("./static/index.html")
	if err != nil {
		log.Printf("Failed to load template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, postList); err != nil {
		log.Printf("Failed to render template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (app *application) basicAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return early if auth is disabled
		if !app.enabledModules["auth"] {
			next.ServeHTTP(w, r)
			return
		}

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

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if app.enabledModules["database"] {
		// Ping database
		err = app.db.Ping(ctx)
		if err != nil {
			log.Printf("Database failure: %v", err)
			http.Error(w, "Database failure", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)

	_, err = w.Write([]byte("OK"))
	if err != nil {
		log.Printf("Failed to write the data to the connection: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func enforceJSONMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")

		if contentType != "" {
			mt, _, err := mime.ParseMediaType(contentType)
			if err != nil {
				http.Error(w, "Malformed Content-Type header", http.StatusBadRequest)
				return
			}

			if mt != "application/json" {
				http.Error(w, "Content-Type header must be application/json", http.StatusUnsupportedMediaType)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// func httpsRedirectMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if r.URL.Scheme != "https" {
// 			if r.Method != http.MethodGet && r.Method != http.MethodHead {
// 				http.Error(w, "HTTPS is required for this request method.", http.StatusUpgradeRequired)
// 				return
// 			}

// 			httpsURL := "https://" + r.Host + r.URL.Path
// 			if r.URL.RawQuery != "" {
// 				httpsURL += "?" + r.URL.RawQuery
// 			}
// 			http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
// 			return
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }

