package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"jlrdi/internal/config"
	"jlrdi/internal/db"
	"jlrdi/internal/httpserver"
	"jlrdi/internal/rag"
	"jlrdi/internal/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Load configuration
	cfg := config.Load()
	log.Printf("Starting JLR Document Intelligence Backend (env=%s, port=%s)", cfg.Env, cfg.Port)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database connection with retry
	pgPool, err := initDatabase(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer pgPool.Close()

	// Initialize S3 signer
	signer, err := storage.NewSignerFromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize S3 signer: %v", err)
	}

	// Seed sample data in development mode
	if cfg.Env == "dev" {
		if err := seedSampleData(ctx, cfg.QdrantURL); err != nil {
			log.Printf("Warning: Failed to seed sample data: %v", err)
		}
	}

	// Create HTTP server
	server := createHTTPServer(cfg, pgPool, signer)

	// Start server in goroutine
	go func() {
		log.Printf("API server listening on :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

func initDatabase(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	var pgPool *pgxpool.Pool
	var err error

	// Retry database connection with exponential backoff
	for i := 0; i < 30; i++ {
		pgPool, err = db.Connect(ctx, cfg.DBURL)
		if err == nil {
			log.Printf("Successfully connected to database")
			break
		}

		backoff := time.Duration(i+1) * 2 * time.Second
		log.Printf("Waiting for database connection... (attempt %d/30): %v, retrying in %v", i+1, err, backoff)
		time.Sleep(backoff)
	}

	if err != nil {
		return nil, err
	}

	// Run database migrations
	if err := db.RunMigrations(ctx, pgPool); err != nil {
		log.Printf("Warning: Failed to run migrations: %v", err)
	}

	return pgPool, nil
}

func createHTTPServer(cfg config.Config, pgPool *pgxpool.Pool, signer *storage.Signer) *http.Server {
	router := httpserver.NewRouter(httpserver.Deps{
		DB:          pgPool,
		Signer:      signer,
		Config:      cfg,
		QdrantURL:   cfg.QdrantURL,
		HTTPTimeout: 15 * time.Second,
	})

	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 15 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func seedSampleData(ctx context.Context, qdrantURL string) error {
	log.Println("Seeding sample data for development...")

	seeder := rag.NewSeederService(qdrantURL)
	return seeder.SeedSampleData(ctx)
}
