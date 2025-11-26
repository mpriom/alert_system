package handler

import (
	"context"
	"encoding/json"
	"log"
	"math/rand/v2"
	"net/http"
	"time"

	"mock-alerts-api/internal/service"
)

type AlertsHandler struct {
	generator   *service.AlertGenerator
	failureRate float64
}

type AlertsResponse struct {
	Alerts []service.ExternalAlert `json:"alerts"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewAlertsHandler(gen *service.AlertGenerator, failureRate float64) *AlertsHandler {
	return &AlertsHandler{
		generator:   gen,
		failureRate: failureRate,
	}
}

// GetAlerts handles GET /alerts with optional ?since= parameter
func (h *AlertsHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed. Use GET.")
		return
	}

	// Simulate random failures
	if h.shouldFail() {
		log.Printf("[MOCK API] Simulated failure (failure rate: %.0f%%)", h.failureRate*100)
		h.writeError(w, http.StatusInternalServerError, "Service temporarily unavailable")
		return
	}

	ctx := r.Context()
	sinceParam := r.URL.Query().Get("since")

	var alerts []service.ExternalAlert
	var err error

	if sinceParam != "" {
		alerts, err = h.getAlertsSince(ctx, sinceParam)
	} else {
		alerts, err = h.getAllAlerts(ctx)
	}

	if err != nil {
		log.Printf("[MOCK API] Error fetching alerts: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to fetch alerts")
		return
	}

	h.writeJSON(w, http.StatusOK, AlertsResponse{Alerts: alerts})
	log.Printf("[MOCK API] Successfully returned %d alerts", len(alerts))
}

func (h *AlertsHandler) getAlertsSince(ctx context.Context, sinceParam string) ([]service.ExternalAlert, error) {
	since, err := time.Parse(time.RFC3339, sinceParam)
	if err != nil {
		return nil, err
	}

	log.Printf("[MOCK API] Fetching alerts since: %s", since.Format(time.RFC3339))
	return h.generator.GetAlertsSince(ctx, since)
}

func (h *AlertsHandler) getAllAlerts(ctx context.Context) ([]service.ExternalAlert, error) {
	log.Printf("[MOCK API] Fetching all alerts")
	return h.generator.GetAllAlerts(ctx)
}

func (h *AlertsHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[MOCK API] Error encoding JSON response: %v", err)
	}
}

func (h *AlertsHandler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, ErrorResponse{Error: message})
}

func (h *AlertsHandler) shouldFail() bool {
	return rand.Float64() < h.failureRate
}

func (h *AlertsHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
