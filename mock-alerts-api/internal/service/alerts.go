package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

var (
	// ValidSeverities defines allowed severity levels
	ValidSeverities = map[string]bool{
		"low":      true,
		"medium":   true,
		"high":     true,
		"critical": true,
	}

	// ValidSources defines the 10 allowed alert sources
	ValidSources = map[string]bool{
		"siem-1":                true,
		"siem-2":                true,
		"firewall":              true,
		"ids":                   true,
		"antivirus":             true,
		"endpoint":              true,
		"cloud-security":        true,
		"email-gateway":         true,
		"network-monitor":       true,
		"vulnerability-scanner": true,
	}
)

// ExternalAlert represents an alert from the third-party system
type ExternalAlert struct {
	Source      string    `json:"source"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// AlertGenerator generates alerts from the external_alerts table
type AlertGenerator struct {
	db *sql.DB
}

// NewAlertGenerator creates a new alert generator
func NewAlertGenerator(db *sql.DB) *AlertGenerator {
	return &AlertGenerator{db: db}
}

// GetAlertsSince fetches alerts created after the given timestamp
func (g *AlertGenerator) GetAlertsSince(ctx context.Context, since time.Time) ([]ExternalAlert, error) {
	query := `
		SELECT source, severity, description, created_at
		FROM external_alerts
		WHERE created_at > $1
		ORDER BY created_at ASC
	`

	rows, err := g.db.QueryContext(ctx, query, since)
	if err != nil {
		return nil, fmt.Errorf("error querying external alerts: %w", err)
	}
	defer rows.Close()

	return g.scanAndFilterAlerts(rows)
}

// GetAllAlerts fetches all alerts (limited to 100)
func (g *AlertGenerator) GetAllAlerts(ctx context.Context) ([]ExternalAlert, error) {
	query := `
		SELECT source, severity, description, created_at
		FROM external_alerts
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := g.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying all external alerts: %w", err)
	}
	defer rows.Close()

	return g.scanAndFilterAlerts(rows)
}

func (g *AlertGenerator) scanAndFilterAlerts(rows *sql.Rows) ([]ExternalAlert, error) {
	var alerts []ExternalAlert

	for rows.Next() {
		var alert ExternalAlert
		err := rows.Scan(
			&alert.Source,
			&alert.Severity,
			&alert.Description,
			&alert.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning alert: %w", err)
		}

		// Only include alerts with valid source and severity
		if IsValidSource(alert.Source) && IsValidSeverity(alert.Severity) {
			alerts = append(alerts, alert)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alerts: %w", err)
	}

	return alerts, nil
}

// IsValidSeverity checks if severity is valid
func IsValidSeverity(severity string) bool {
	return ValidSeverities[severity]
}

// IsValidSource checks if source is valid
func IsValidSource(source string) bool {
	return ValidSources[source]
}

// GetValidSources returns list of valid sources
func GetValidSources() []string {
	sources := make([]string, 0, len(ValidSources))
	for s := range ValidSources {
		sources = append(sources, s)
	}
	return sources
}

// GetValidSeverities returns list of valid severities
func GetValidSeverities() []string {
	severities := make([]string, 0, len(ValidSeverities))
	for s := range ValidSeverities {
		severities = append(severities, s)
	}
	return severities
}
