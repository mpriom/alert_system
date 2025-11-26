package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"censys_alert_system/external"
	"censys_alert_system/internal/models"
)

type AlertService struct {
	mockAPIClient APIClientInterface
	storage       AlertStorageInterface
}

func NewAlertService(storage AlertStorageInterface, apiClient APIClientInterface) *AlertService {
	return &AlertService{
		storage:       storage,
		mockAPIClient: apiClient,
	}
}

// GetAlerts retrieves all alerts through the service layer
func (s *AlertService) GetAlerts(ctx context.Context) ([]models.Alert, error) {
	alerts, err := s.storage.GetAlerts(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: error getting alerts: %w", err)
	}

	return alerts, nil
}

// GetAlertByID retrieves a single alert by ID through the service layer
func (s *AlertService) GetAlertByID(ctx context.Context, id string) (*models.Alert, error) {
	alert, err := s.storage.GetAlertByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service: error getting alert: %w", err)
	}

	return alert, nil
}

// GetAlertsByDays retrieves alerts from the last X days through the service layer
func (s *AlertService) GetAlertsByDays(ctx context.Context, days int) ([]models.Alert, error) {
	if days <= 0 {
		return nil, fmt.Errorf("service: days must be greater than 0")
	}

	alerts, err := s.storage.GetAlertsByDays(ctx, days)
	if err != nil {
		return nil, fmt.Errorf("service: error getting alerts by days: %w", err)
	}

	return alerts, nil
}

// PerformSync fetches alerts from mock API, enriches them, and stores them
func (s *AlertService) PerformSync(ctx context.Context) error {
	log.Println("[SYNC] Starting sync process...")

	// Health check
	if err := s.mockAPIClient.CheckHealth(ctx); err != nil {
		log.Printf("[SYNC] Health check failed: %v, proceeding anyway (retries will handle failures)", err)
	}

	// Get last sync time from database
	lastSync, err := s.storage.GetLastSyncTime(ctx)
	if err != nil {
		log.Printf("[SYNC] Warning: Could not get last sync time: %v. Fetching all alerts.", err)
		lastSync = time.Time{}
	}

	log.Printf("[SYNC] Last sync time: %s", lastSync.Format(time.RFC3339))

	// Fetch alerts from mock API
	var externalAlerts []external.ExternalAlert
	if lastSync.IsZero() {
		log.Println("[SYNC] Fetching all alerts (first sync)")
		externalAlerts, err = s.mockAPIClient.FetchAllAlerts(ctx)
	} else {
		log.Printf("[SYNC] Fetching alerts since: %s", lastSync.Format(time.RFC3339))
		externalAlerts, err = s.mockAPIClient.FetchAlertsSince(ctx, lastSync)
	}

	if err != nil {
		return fmt.Errorf("failed to fetch alerts from mock API: %w", err)
	}

	log.Printf("[SYNC] Fetched %d alerts from mock API", len(externalAlerts))

	if len(externalAlerts) == 0 {
		log.Println("[SYNC] No new alerts to sync")
		return nil
	}

	// Process and store each alert
	successCount := 0
	for _, extAlert := range externalAlerts {
		if ctx.Err() != nil {
			log.Printf("[SYNC] Sync cancelled after processing %d alerts", successCount)
			return ctx.Err()
		}

		wholeEventJSON, err := json.Marshal(map[string]interface{}{
			"source":      extAlert.Source,
			"severity":    extAlert.Severity,
			"description": extAlert.Description,
			"created_at":  extAlert.CreatedAt,
			"synced_at":   time.Now(),
		})
		if err != nil {
			log.Printf("[SYNC] Warning: Failed to marshal whole_event for alert: %v", err)
			wholeEventJSON = []byte("{}")
		}

		err = s.storage.CreateAlert(
			ctx,
			extAlert.Source,
			extAlert.Severity,
			extAlert.Description,
			wholeEventJSON,
			getRandomEnrichmentType(),
			generateRandomIP(),
			extAlert.CreatedAt,
		)

		if err != nil {
			log.Printf("[SYNC] Error storing alert: %v", err)
			continue
		}

		successCount++
	}

	// Update last sync time
	if err := s.storage.UpdateLastSyncTime(ctx, time.Now()); err != nil {
		log.Printf("[SYNC] Warning: Failed to update last sync time: %v", err)
	}

	log.Printf("[SYNC] Successfully synced %d/%d alerts", successCount, len(externalAlerts))
	return nil
}

func generateRandomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		rand.Intn(256),
		rand.Intn(256),
		rand.Intn(256),
		rand.Intn(256),
	)
}

func getRandomEnrichmentType() string {
	alertTypes := []string{
		"geo_location",
		"threat_intel",
		"user_context",
		"network_analysis",
		"behavioral_analysis",
	}

	return alertTypes[rand.Intn(len(alertTypes))]
}
