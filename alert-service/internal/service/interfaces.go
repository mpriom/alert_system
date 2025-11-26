package service

import (
	"context"
	"time"

	"censys_alert_system/external"
	"censys_alert_system/internal/models"
)

// AlertStorageInterface defines the contract for alert storage operations.
// Implemented by storage.AlertStorage
//
//go:generate mockery --name=AlertStorageInterface --output=./mocks --outpkg=mocks
type AlertStorageInterface interface {
	GetAlerts(ctx context.Context) ([]models.Alert, error)
	GetAlertByID(ctx context.Context, id string) (*models.Alert, error)
	GetAlertsByDays(ctx context.Context, days int) ([]models.Alert, error)
	CreateAlert(ctx context.Context, source, severity, description string, wholeEvent []byte, enrichmentType, ipAddress string, createdAt time.Time) error
	GetLastSyncTime(ctx context.Context) (time.Time, error)
	UpdateLastSyncTime(ctx context.Context, t time.Time) error
}

// APIClientInterface defines the contract for external API operations.
// Implemented by external.MockAPIClient
//
//go:generate mockery --name=APIClientInterface --output=./mocks --outpkg=mocks
type APIClientInterface interface {
	CheckHealth(ctx context.Context) error
	FetchAllAlerts(ctx context.Context) ([]external.ExternalAlert, error)
	FetchAlertsSince(ctx context.Context, since time.Time) ([]external.ExternalAlert, error)
}