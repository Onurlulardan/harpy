package billingv1

import (
	"encoding/json"
	"testing"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestLegacyEventPresentsExactCompatibilityMembers(t *testing.T) {
	subscriptionID := "11111111-1111-4111-8111-111111111111"
	metadata := json.RawMessage(`{"invoice":"INV-1"}`)
	created := "2026-07-17T08:00:00.123400Z"
	view := billingapi.LegacyEventView{
		ID:             "22222222-2222-4222-8222-222222222222",
		TenantID:       "33333333-3333-4333-8333-333333333333",
		SubscriptionID: &subscriptionID, EventType: "payment_succeeded",
		AmountCents: 1250, Currency: "TRY", Description: "paid",
		Metadata: &metadata, CreatedAt: created,
	}
	wire, err := legacyEvent(view)
	if err != nil {
		t.Fatal(err)
	}
	document, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	createdAt, _ := time.Parse(utcMicrosecondLayout, created)
	expected, _ := json.Marshal(v1EventReference{
		ID: view.ID, TenantID: view.TenantID, SubscriptionID: &subscriptionID,
		EventType: view.EventType, AmountCents: 1250, Currency: "TRY",
		Description: "paid", Metadata: metadata, CreatedAt: createdAt,
	})
	if string(document) != string(expected) {
		t.Fatalf("wire=%s want=%s", document, expected)
	}
}

func TestLegacyEventOmitsExactOptionalMembers(t *testing.T) {
	wire, err := legacyEvent(billingapi.LegacyEventView{
		ID:        "22222222-2222-4222-8222-222222222222",
		TenantID:  "33333333-3333-4333-8333-333333333333",
		EventType: "payment_failed", Currency: "usd",
		CreatedAt: "2026-07-17T08:00:00.000000Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	document, _ := json.Marshal(wire)
	var members map[string]json.RawMessage
	_ = json.Unmarshal(document, &members)
	if len(members) != 6 || members["subscription_id"] != nil ||
		members["description"] != nil || members["metadata"] != nil {
		t.Fatalf("optional event wire = %s", document)
	}
}

func TestLegacyEventRejectsInvalidEvidence(t *testing.T) {
	_, err := legacyEvent(billingapi.LegacyEventView{
		ID:        "22222222-2222-4222-8222-222222222222",
		TenantID:  "33333333-3333-4333-8333-333333333333",
		EventType: "payment_succeeded", Currency: "TOOLONG",
		CreatedAt: "2026-07-17T08:00:00.000000Z",
	})
	if err == nil {
		t.Fatal("invalid event currency was accepted")
	}
}

type v1EventReference struct {
	ID             string          `json:"id"`
	TenantID       string          `json:"tenant_id"`
	SubscriptionID *string         `json:"subscription_id,omitempty"`
	EventType      string          `json:"event_type"`
	AmountCents    int64           `json:"amount_cents"`
	Currency       string          `json:"currency"`
	Description    string          `json:"description,omitempty"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}
