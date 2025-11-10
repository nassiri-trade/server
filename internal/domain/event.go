package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

// CalendarEvent represents a single ForexFactory calendar entry in the domain layer.
type CalendarEvent struct {
	Hash      string
	Date      time.Time
	Currency  string
	Impact    string
	Detail    string
	Actual    string
	Forecast  string
	Previous  string
	SourceURL string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IdentityDigest builds a deterministic hash for the event using immutable identity fields.
// The digest intentionally omits Actual/Forecast/Previous so updates can be upserted.
func IdentityDigest(date time.Time, currency, detail string) string {
	parts := []string{
		date.UTC().Format(time.RFC3339),
		strings.TrimSpace(strings.ToUpper(currency)),
		strings.TrimSpace(strings.ToLower(detail)),
	}

	h := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(h[:])
}

// WithHash ensures the Hash field is populated based on the event identity.
func (e CalendarEvent) WithHash() CalendarEvent {
	if e.Hash != "" {
		return e
	}

	e.Hash = IdentityDigest(e.Date, e.Currency, e.Detail)
	return e
}

