package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"s3-presigner/internal/config"
	"s3-presigner/internal/httpserver"
	"s3-presigner/internal/storage"
)

const shutdownTimeout = 10 * time.Second

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Initialize config
	cfg := config.New()

	// Validate AWS credentials
	if err := storage.ValidateAWSCredentials(cfg); err != nil {
		log.Fatalf("AWS credentials validation failed: %v", err)
	}
	log.Println("AWS credentials verified successfully")

	// Setup and start server
	server := httpserver.NewServer(cfg)

	// Graceful shutdown setup
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Could not gracefully shutdown the server: %v", err)
		}
	}()

	log.Printf("Server starting on port %s...", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Error starting server: %v", err)
	}

	log.Println("Server stopped")
}
