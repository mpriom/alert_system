package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"censys_alert_system/internal/models"
)

type AlertStorage struct {
	db *sql.DB
}

func NewAlertStorage(db *sql.DB) *AlertStorage {
	return &AlertStorage{db: db}
}

// CreateAlert inserts a new alert into the database with enrichment
func (s *AlertStorage) CreateAlert(ctx context.Context, source, severity, description string, wholeEvent []byte, enrichmentType, ipAddress string, createdAt time.Time) error {
	query := `
		INSERT INTO alerts (source, severity, description, whole_event, enrichment_type, ip_address, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := s.db.ExecContext(ctx, query, source, severity, description, wholeEvent, enrichmentType, ipAddress, createdAt)
	if err != nil {
		return fmt.Errorf("error creating alert: %w", err)
	}

	return nil
}

// GetAlerts retrieves all alerts from the database
func (s *AlertStorage) GetAlerts(ctx context.Context) ([]models.Alert, error) {
	query := `
		SELECT id, source, severity, description, whole_event, enrichment_type, ip_address, created_at
		FROM alerts
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying alerts: %w", err)
	}
	defer rows.Close()

	var alerts []models.Alert
	for rows.Next() {
		var alert models.Alert
		err := rows.Scan(
			&alert.ID,
			&alert.Source,
			&alert.Severity,
			&alert.Description,
			&alert.WholeEvent,
			&alert.EnrichmentType,
			&alert.IPAddress,
			&alert.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning alert: %w", err)
		}
		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alerts: %w", err)
	}

	return alerts, nil
}

// GetAlertByID retrieves a single alert by ID
func (s *AlertStorage) GetAlertByID(ctx context.Context, id string) (*models.Alert, error) {
	query := `
		SELECT id, source, severity, description, whole_event, enrichment_type, ip_address, created_at
		FROM alerts
		WHERE id = $1
	`

	var alert models.Alert
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&alert.ID,
		&alert.Source,
		&alert.Severity,
		&alert.Description,
		&alert.WholeEvent,
		&alert.EnrichmentType,
		&alert.IPAddress,
		&alert.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("alert not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error querying alert: %w", err)
	}

	return &alert, nil
}

// GetAlertsByDays retrieves alerts from the last X days
func (s *AlertStorage) GetAlertsByDays(ctx context.Context, days int) ([]models.Alert, error) {
	query := `
		SELECT id, source, severity, description, whole_event, enrichment_type, ip_address, created_at
		FROM alerts
		WHERE created_at >= NOW() - INTERVAL '1 day' * $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, days)
	if err != nil {
		return nil, fmt.Errorf("error querying alerts by days: %w", err)
	}
	defer rows.Close()

	var alerts []models.Alert
	for rows.Next() {
		var alert models.Alert
		err := rows.Scan(
			&alert.ID,
			&alert.Source,
			&alert.Severity,
			&alert.Description,
			&alert.WholeEvent,
			&alert.EnrichmentType,
			&alert.IPAddress,
			&alert.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning alert: %w", err)
		}
		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alerts: %w", err)
	}

	return alerts, nil
}

// GetLastSyncTime retrieves the last sync timestamp
func (s *AlertStorage) GetLastSyncTime(ctx context.Context) (time.Time, error) {
	query := `
		SELECT MAX(created_at) FROM alerts
	`

	var lastSync sql.NullTime
	err := s.db.QueryRowContext(ctx, query).Scan(&lastSync)
	if err != nil {
		return time.Time{}, fmt.Errorf("error querying last sync time: %w", err)
	}

	if !lastSync.Valid {
		return time.Time{}, nil
	}

	return lastSync.Time, nil
}

// UpdateLastSyncTime - In this implementation, sync time is implicit (most recent alert)
// This is a no-op but kept for interface consistency
func (s *AlertStorage) UpdateLastSyncTime(ctx context.Context, t time.Time) error {
	return nil
}
