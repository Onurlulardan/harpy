package domain

import (
	"bytes"
	"testing"
	"time"
)

func deliveryFixture(t *testing.T) Delivery {
	t.Helper()
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	id, _ := NewV7(now, bytes.NewReader(bytes.Repeat([]byte{3}, 10)))
	message, _ := ParseID("11111111-1111-4111-8111-111111111111")
	value, err := NewDelivery(Delivery{ID: id, MessageID: message, Channel: ChannelEmail,
		State: DeliveryPending, MaximumAttempts: 8, NotBefore: now, Revision: 1})
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestOutcomeUnknownCannotRetryOrReplay(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	claimed, err := deliveryFixture(t).Claim("worker-a", now)
	if err != nil {
		t.Fatal(err)
	}
	failed, err := claimed.Fail("worker-a", claimed.LeaseGeneration,
		"NOTIFICATION_OUTCOME_UNKNOWN", true)
	if err != nil || !failed.OutcomeUnknown {
		t.Fatalf("Fail() = %+v, %v", failed, err)
	}
	if _, err = failed.Retry(now, now.Add(time.Minute)); err != ErrDeliveryConflict {
		t.Fatalf("Retry() error = %v", err)
	}
}

func TestDeliveryLeaseFencesStaleWorker(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	claimed, _ := deliveryFixture(t).Claim("worker-a", now)
	if _, err := claimed.Accept("worker-b", claimed.LeaseGeneration, "ACCEPTED"); err != ErrDeliveryConflict {
		t.Fatalf("stale worker error = %v", err)
	}
}

func TestProviderReceiptDistinguishesAcceptedDeliveredAndBounced(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	claimed, _ := deliveryFixture(t).Claim("worker-a", now)
	accepted, err := claimed.Accept("worker-a", claimed.LeaseGeneration, "ACCEPTED")
	if err != nil || accepted.State != DeliveryAccepted {
		t.Fatalf("Accept() = %+v, %v", accepted, err)
	}
	delivered, err := accepted.RecordProviderReceipt(true, "DELIVERED")
	if err != nil || delivered.State != DeliveryDelivered {
		t.Fatalf("RecordProviderReceipt() = %+v, %v", delivered, err)
	}
}
