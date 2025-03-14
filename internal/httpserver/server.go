package httpserver

import (
	"log"
	"net/http"
	"s3-presigner/internal/config"
	"s3-presigner/internal/storage"
	"time"
)

// LoggingMiddleware wraps all handlers by default to log request details
type loggingHandler struct {
	handler http.Handler
}

func (l loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	log.Printf("Started %s %s", r.Method, r.URL.Path)
	l.handler.ServeHTTP(w, r)
	log.Printf("Completed %s %s in %v", r.Method, r.URL.Path, time.Since(start))
}

func NewServer(cfg *config.Config) *http.Server {
	s, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/presign/get", GetPresignHandler(s))
	mux.HandleFunc("/presign/delete", DeletePresignHandler(s))
	mux.HandleFunc("/presign/put", PutPresignHandler(s))

	// Wrap the entire mux with logging middleware
	return &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      loggingHandler{handler: mux},
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
