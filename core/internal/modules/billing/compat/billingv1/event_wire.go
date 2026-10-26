package billingv1

import (
	"encoding/json"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

type eventWire struct {
	ID             string          `json:"id"`
	TenantID       string          `json:"tenant_id"`
	SubscriptionID *string         `json:"subscription_id,omitempty"`
	EventType      string          `json:"event_type"`
	AmountCents    int64           `json:"amount_cents"`
	Currency       string          `json:"currency"`
	Description    string          `json:"description,omitempty"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
	CreatedAt      string          `json:"created_at"`
}

func legacyEvent(view billingapi.LegacyEventView) (eventWire, error) {
	if view.EventType == "" || !validEventCurrency(view.Currency) {
		return eventWire{}, ErrSnapshotRequired
	}
	for _, identifier := range []string{view.ID, view.TenantID} {
		parsed, err := uuid.Parse(identifier)
		if err != nil || parsed == uuid.Nil || parsed.String() != identifier {
			return eventWire{}, ErrSnapshotRequired
		}
	}
	if view.SubscriptionID != nil {
		identifier, err := uuid.Parse(*view.SubscriptionID)
		if err != nil || identifier == uuid.Nil || identifier.String() != *view.SubscriptionID {
			return eventWire{}, ErrSnapshotRequired
		}
	}
	createdAt, err := v1WireInstant(view.CreatedAt)
	if err != nil {
		return eventWire{}, ErrSnapshotRequired
	}
	var metadata json.RawMessage
	if view.Metadata != nil {
		if !json.Valid(*view.Metadata) {
			return eventWire{}, ErrSnapshotRequired
		}
		metadata = append(json.RawMessage(nil), (*view.Metadata)...)
	}
	return eventWire{
		ID: view.ID, TenantID: view.TenantID,
		SubscriptionID: view.SubscriptionID, EventType: view.EventType,
		AmountCents: view.AmountCents, Currency: view.Currency,
		Description: view.Description,
		Metadata:    metadata, CreatedAt: createdAt,
	}, nil
}

func validEventCurrency(value string) bool {
	count := 0
	for range value {
		count++
	}
	return count > 0 && count <= 3
}
