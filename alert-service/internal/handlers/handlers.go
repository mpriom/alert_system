package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"censys_alert_system/internal/models"
	"censys_alert_system/internal/service"
)

type AlertHandler struct {
	alertService *service.AlertService
	syncTimeout  time.Duration
}

type AlertsResponse struct {
	Alerts []models.Alert `json:"alerts"`
}

type SingleAlertResponse struct {
	Alert *models.Alert `json:"alert"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SyncResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func NewAlertHandler(alertService *service.AlertService) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
		syncTimeout:  5 * time.Minute,
	}
}

// writeJSON writes a JSON response with the given status code
func (h *AlertHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[HANDLER] Error encoding JSON response: %v", err)
	}
}

// writeError writes a JSON error response
func (h *AlertHandler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, ErrorResponse{Error: message})
}

// GetAlerts handles GET /alerts with optional query parameters
// Query params:
//   - id: Get a specific alert by ID
//   - days: Get alerts from the last N days
//   - (none): Get all alerts
func (h *AlertHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed. Use GET.")
		return
	}

	ctx := r.Context()
	idParam := r.URL.Query().Get("id")
	daysParam := r.URL.Query().Get("days")

	if idParam != "" && daysParam != "" {
		h.writeError(w, http.StatusBadRequest, "Cannot specify both 'id' and 'days' parameters at the same time")
		return
	}

	switch {
	case idParam != "":
		h.getAlertByID(ctx, w, idParam)
	case daysParam != "":
		h.getAlertsByDays(ctx, w, daysParam)
	default:
		h.getAllAlerts(ctx, w)
	}
}

// getAlertByID retrieves a single alert by its ID
func (h *AlertHandler) getAlertByID(ctx context.Context, w http.ResponseWriter, id string) {
	alert, err := h.alertService.GetAlertByID(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Alert not found")
		return
	}

	h.writeJSON(w, http.StatusOK, SingleAlertResponse{Alert: alert})
}

// getAlertsByDays retrieves alerts from the last N days
func (h *AlertHandler) getAlertsByDays(ctx context.Context, w http.ResponseWriter, daysParam string) {
	days, err := strconv.Atoi(daysParam)
	if err != nil || days <= 0 {
		h.writeError(w, http.StatusBadRequest, "Invalid 'days' parameter. Must be a positive integer")
		return
	}

	alerts, err := h.alertService.GetAlertsByDays(ctx, days)
	if err != nil {
		log.Printf("[HANDLER] Error getting alerts by days: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to retrieve alerts")
		return
	}

	h.writeJSON(w, http.StatusOK, AlertsResponse{Alerts: alerts})
}

// getAllAlerts retrieves all alerts
func (h *AlertHandler) getAllAlerts(ctx context.Context, w http.ResponseWriter) {
	alerts, err := h.alertService.GetAlerts(ctx)
	if err != nil {
		log.Printf("[HANDLER] Error getting all alerts: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to retrieve alerts")
		return
	}

	h.writeJSON(w, http.StatusOK, AlertsResponse{Alerts: alerts})
}

// TriggerSync handles POST /sync to manually trigger a sync
func (h *AlertHandler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed. Use POST.")
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), h.syncTimeout)
		defer cancel()

		if err := h.alertService.PerformSync(ctx); err != nil {
			log.Printf("[SYNC] Error during manual sync: %v", err)
		} else {
			log.Printf("[SYNC] Manual sync completed successfully")
		}
	}()

	h.writeJSON(w, http.StatusAccepted, SyncResponse{
		Message: "Sync triggered successfully",
		Status:  "pending",
	})
}

func (h *AlertHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
