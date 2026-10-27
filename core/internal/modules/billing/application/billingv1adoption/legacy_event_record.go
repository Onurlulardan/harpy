package billingv1adoptionapp

import (
	"encoding/json"
	"time"
)

type LegacyEventRecord struct {
	SourceID             string          `json:"source_id"`
	ID                   string          `json:"id"`
	TenantID             string          `json:"tenant_id"`
	SubscriptionID       *string         `json:"subscription_id"`
	SourceSubscriptionID *string         `json:"source_subscription_id"`
	EventType            string          `json:"event_type"`
	AmountCents          int64           `json:"amount_cents"`
	Currency             string          `json:"currency"`
	Description          string          `json:"description"`
	Metadata             json.RawMessage `json:"metadata,omitempty"`
	CreatedAt            time.Time       `json:"created_at"`
	SourceSHA256         string          `json:"source_sha256"`
}
