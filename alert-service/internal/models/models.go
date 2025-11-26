package models

import "time"

type Alert struct {
	ID             string    `json:"id"`
	Source         string    `json:"source"`
	Severity       string    `json:"severity"`
	Description    string    `json:"description"`
	WholeEvent     []byte    `json:"whole_event"`
	EnrichmentType *string   `json:"enrichment_type"`
	IPAddress      *string   `json:"ip_address"`
	CreatedAt      time.Time `json:"created_at"`
}
