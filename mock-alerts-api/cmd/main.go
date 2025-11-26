package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mock-alerts-api/config"
	"mock-alerts-api/internal/handler"
	"mock-alerts-api/internal/service"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	log.Printf("Mock Alerts API Configuration:")
	log.Printf("  Port: %s", cfg.Port)
	log.Printf("  Failure Rate: %.0f%%", cfg.FailureRate*100)
	log.Printf("  Database: %s@%s:%s/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)

	// Initialize database connection
	db, err := config.NewDB(cfg.GetDBConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Successfully connected to database")

	// Initialize layers
	alertGen := service.NewAlertGenerator(db)
	alertsHandler := handler.NewAlertsHandler(alertGen, cfg.FailureRate)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/alerts", alertsHandler.GetAlerts)
	mux.HandleFunc("/health", alertsHandler.HealthHandler)

	// Create server with timeouts
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Mock Alerts API starting on http://localhost%s", server.Addr)
		log.Printf("Endpoints:")
		log.Printf("  GET  /alerts  - Fetch alerts (optional ?since=<ISO8601>)")
		log.Printf("  GET  /health  - Health check")
		log.Printf("Simulating %.0f%% random failures on /alerts endpoint", cfg.FailureRate*100)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
