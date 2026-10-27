package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestDecodeSubscriptionFactRequiresCanonicalTypedSnapshot(t *testing.T) {
	subscriptionID := uuid.New()
	start := time.Date(2026, 7, 17, 9, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	document := []byte(fmt.Sprintf(
		`{"boundary":null,"grace_ends_at":null,"period_end":%q,"period_start":%q,"state":"active","subscription_id":%q,"trial_ends_at":null}`,
		end.Format(time.RFC3339Nano), start.Format(time.RFC3339Nano), subscriptionID,
	))
	fact, err := decodeSubscriptionFact(document)
	if err != nil || fact.subscriptionID != subscriptionID || fact.state != "active" ||
		fact.period == nil || !fact.period.Start().Equal(start) ||
		!fact.period.End().Equal(end) || fact.boundary != nil {
		t.Fatalf("fact=%+v error=%v", fact, err)
	}
	invalid := [][]byte{
		[]byte(fmt.Sprintf(
			`{"state":"active","subscription_id":%q}`, subscriptionID,
		)),
		[]byte(fmt.Sprintf(
			`{ "boundary":null,"grace_ends_at":null,"period_end":%q,"period_start":%q,"state":"active","subscription_id":%q,"trial_ends_at":null}`,
			end.Format(time.RFC3339Nano), start.Format(time.RFC3339Nano), subscriptionID,
		)),
		[]byte(fmt.Sprintf(
			`{"boundary":null,"grace_ends_at":null,"period_end":null,"period_start":%q,"state":"active","subscription_id":%q,"trial_ends_at":null}`,
			start.Format(time.RFC3339Nano), subscriptionID,
		)),
	}
	for _, document := range invalid {
		if _, err := decodeSubscriptionFact(document); err == nil {
			t.Fatalf("accepted invalid subscription fact: %s", document)
		}
	}
}

func TestDecodeSubscriptionFactValidatesBoundaryEvidence(t *testing.T) {
	subscriptionID := uuid.New()
	jobID, _ := uuid.NewV7()
	start := time.Date(2026, 7, 17, 9, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	document := []byte(fmt.Sprintf(
		`{"boundary":{"due_at":%q,"expected_revision":4,"job_id":%q,"kind":"period_end","outcome":"paid"},"grace_ends_at":null,"period_end":%q,"period_start":%q,"state":"active","subscription_id":%q,"trial_ends_at":null}`,
		start.Format(time.RFC3339Nano), jobID, end.Format(time.RFC3339Nano),
		start.Format(time.RFC3339Nano), subscriptionID,
	))
	fact, err := decodeSubscriptionFact(document)
	if err != nil || fact.boundary == nil || fact.boundary.jobID != jobID ||
		fact.boundary.outcome != "paid" {
		t.Fatalf("boundary=%+v error=%v", fact.boundary, err)
	}
	invalid := []byte(fmt.Sprintf(
		`{"boundary":{"due_at":%q,"expected_revision":4,"job_id":%q,"kind":"period_end","outcome":"unpaid"},"grace_ends_at":null,"period_end":%q,"period_start":%q,"state":"active","subscription_id":%q,"trial_ends_at":null}`,
		start.Format(time.RFC3339Nano), jobID, end.Format(time.RFC3339Nano),
		start.Format(time.RFC3339Nano), subscriptionID,
	))
	if _, err := decodeSubscriptionFact(invalid); err == nil {
		t.Fatal("accepted unpaid evidence for active subscription")
	}
}
