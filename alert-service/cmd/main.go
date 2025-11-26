package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"censys_alert_system/config"
	"censys_alert_system/external"
	"censys_alert_system/internal/handlers"
	"censys_alert_system/internal/service"
	"censys_alert_system/internal/storage"
)

func main() {
	cfg := config.LoadConfig()

	log.Printf("Alert Service Configuration:")
	log.Printf("  Database: %s@%s:%s/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
	log.Printf("  Mock API URL: %s", cfg.MockAPIURL)
	log.Printf("  Sync Interval: %s", cfg.SyncInterval)

	db, err := config.NewDB(cfg.GetDBConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Successfully connected to database")

	alertStorage := storage.NewAlertStorage(db)
	mockAPIClient := external.NewMockAPIClient(cfg.MockAPIURL)
	alertService := service.NewAlertService(alertStorage, mockAPIClient)
	alertHandler := handlers.NewAlertHandler(alertService)

	mux := http.NewServeMux()
	mux.HandleFunc("/alerts", alertHandler.GetAlerts)
	mux.HandleFunc("/sync", alertHandler.TriggerSync)
	mux.HandleFunc("/health", healthHandler)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initial sync
	go runSync(ctx, alertService, "STARTUP")

	// Periodic sync
	go startPeriodicSync(ctx, alertService, cfg.SyncInterval)

	go func() {
		log.Printf("Alert Service starting on http://localhost%s", server.Addr)
		log.Printf("Endpoints:")
		log.Printf("  GET  /alerts  - List alerts (optional: ?id=<uuid> or ?days=<int>)")
		log.Printf("  POST /sync    - Trigger manual sync")
		log.Printf("  GET  /health  - Health check")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

func startPeriodicSync(ctx context.Context, alertService *service.AlertService, interval time.Duration) {
	log.Printf("[SCHEDULER] Starting periodic sync every %s", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[SCHEDULER] Stopping periodic sync")
			return
		case <-ticker.C:
			runSync(ctx, alertService, "SCHEDULED")
		}
	}
}

func runSync(ctx context.Context, alertService *service.AlertService, source string) {
	log.Printf("[%s] Running sync...", source)

	syncCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	if err := alertService.PerformSync(syncCtx); err != nil {
		log.Printf("[%s] Sync failed: %v", source, err)
	} else {
		log.Printf("[%s] Sync completed successfully", source)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
