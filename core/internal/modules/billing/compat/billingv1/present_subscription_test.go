package billingv1

import (
	"encoding/json"
	"testing"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestLegacySubscriptionUsesCanonicalSnapshotMoneyAndTimestamps(t *testing.T) {
	created := "2026-07-01T08:00:00.123400Z"
	updated := "2026-07-17T08:00:00.000000Z"
	legacyID := "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	legacyPlanID := "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	offering := billingapi.OfferingView{
		ID:        "11111111-1111-4111-8111-111111111111",
		VersionID: "22222222-2222-4222-8222-222222222222",
		State:     "retired", Public: false, BillingInterval: "monthly",
		CreatedAt: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC),
		Prices: []billingapi.PriceView{{
			ID:      "33333333-3333-4333-8333-333333333333",
			Cadence: "monthly",
			Amount: billingapi.MoneyView{
				MinorUnits: 10000, Currency: "TRY", RegistryVersion: 1,
			},
		}},
	}
	view := billingapi.SubscriptionView{
		ID:                "44444444-4444-4444-8444-444444444444",
		LegacyID:          legacyID,
		TenantID:          "55555555-5555-4555-8555-555555555555",
		LegacyPlanID:      legacyPlanID,
		OfferingVersionID: offering.VersionID,
		PriceVersionID:    offering.Prices[0].ID,
		State:             "active", PeriodStart: created, PeriodEnd: updated,
		EffectiveMinor: 7500, EffectiveCurrency: "TRY", EffectiveRegistry: 1,
		TrialStartedAt: &created, TrialEndsAt: &updated,
		CancelledAt: &updated, CancellationReason: "requested",
		CreatedAt: created, UpdatedAt: updated,
	}
	wire, err := legacySubscription(view, offering)
	if err != nil {
		t.Fatal(err)
	}
	legacyOffering := offering
	legacyOffering.ID = legacyPlanID
	createdAt, _ := time.Parse(utcMicrosecondLayout, created)
	updatedAt, _ := time.Parse(utcMicrosecondLayout, updated)
	expected := v1SubscriptionReference{
		ID: legacyID, TenantID: view.TenantID, PlanID: legacyPlanID,
		Plan: legacyPlan(legacyOffering), Status: view.State,
		CurrentPeriodStart: createdAt, CurrentPeriodEnd: updatedAt,
		TrialStart: &createdAt, TrialEnd: &updatedAt,
		CancelledAt: &updatedAt, CancellationReason: "requested",
		EffectivePriceCents: 7500, CreatedAt: createdAt, UpdatedAt: updatedAt,
	}
	actualJSON, _ := json.Marshal(wire)
	expectedJSON, _ := json.Marshal(expected)
	if string(actualJSON) != string(expectedJSON) {
		t.Fatalf("wire=%s want=%s", actualJSON, expectedJSON)
	}
}

type v1SubscriptionReference struct {
	ID                  string     `json:"id"`
	TenantID            string     `json:"tenant_id"`
	PlanID              string     `json:"plan_id"`
	Plan                planWire   `json:"plan"`
	Status              string     `json:"status"`
	CurrentPeriodStart  time.Time  `json:"current_period_start"`
	CurrentPeriodEnd    time.Time  `json:"current_period_end"`
	TrialStart          *time.Time `json:"trial_start,omitempty"`
	TrialEnd            *time.Time `json:"trial_end,omitempty"`
	CancelAtPeriodEnd   bool       `json:"cancel_at_period_end"`
	CancelledAt         *time.Time `json:"cancelled_at,omitempty"`
	CancellationReason  string     `json:"cancellation_reason,omitempty"`
	EffectivePriceCents int64      `json:"effective_price_cents"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func TestLegacySubscriptionRejectsMissingEffectiveMoneyEvidence(t *testing.T) {
	offering := billingapi.OfferingView{
		VersionID: "22222222-2222-4222-8222-222222222222",
		Prices: []billingapi.PriceView{{
			ID:     "33333333-3333-4333-8333-333333333333",
			Amount: billingapi.MoneyView{Currency: "USD", RegistryVersion: 1},
		}},
	}
	view := billingapi.SubscriptionView{
		OfferingVersionID: offering.VersionID, PriceVersionID: offering.Prices[0].ID,
		EffectiveCurrency: "TRY", EffectiveRegistry: 1,
		CreatedAt: "2026-07-01T00:00:00.000000Z",
		UpdatedAt: "2026-07-01T00:00:00.000000Z",
	}
	if _, err := legacySubscription(view, offering); err == nil {
		t.Fatal("currency-mismatched effective price was accepted")
	}
}
