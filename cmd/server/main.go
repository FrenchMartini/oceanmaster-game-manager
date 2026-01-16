// Package main is the entry point for the game manager server.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/oceanmining/game-manager/auth"
	"github.com/oceanmining/game-manager/config"
	"github.com/oceanmining/game-manager/database"
	"github.com/oceanmining/game-manager/handlers"
	"github.com/oceanmining/game-manager/redis"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	dbConfig := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}
	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	// Run migrations
	migErr := database.RunMigrations(db)
	if migErr != nil {
		log.Fatalf("Failed to run migrations: %v", migErr)
	}
	log.Println("Database migrations completed")

	// Initialize Redis
	_, err = redis.NewClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Redis connected")

	// Initialize services
	jwtService := auth.NewJWTService(cfg.JWT.Secret, cfg.JWT.ExpiryHours)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, jwtService, cfg.Google)
	healthHandler := handlers.NewHealthHandler()

	// Setup router
	router := mux.NewRouter()

	// Public routes
	router.HandleFunc("/health", healthHandler.Health).Methods("GET")
	router.HandleFunc("/auth/google/login", authHandler.GoogleLogin).Methods("GET")
	router.HandleFunc("/auth/google/callback", authHandler.GoogleCallback).Methods("GET")

	// Start server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
