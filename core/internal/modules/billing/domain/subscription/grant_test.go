package subscription

import (
	"testing"
	"time"

	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
)

func TestPurchasedGrantUsesStateAndHalfOpenBoundary(t *testing.T) {
	start := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	window, err := billingperiod.New(start, end)
	if err != nil {
		t.Fatal(err)
	}
	for _, state := range []State{Trialing, Active, PastDue} {
		active, err := PurchasedGrantActive(state, &window, end.Add(-time.Microsecond))
		if err != nil || !active {
			t.Fatalf("state %s active before boundary = %t, %v", state, active, err)
		}
		active, err = PurchasedGrantActive(state, &window, end)
		if err != nil || active {
			t.Fatalf("state %s active at boundary = %t, %v", state, active, err)
		}
	}
}

func TestPurchasedGrantRejectsMissingOrUnexpectedWindow(t *testing.T) {
	if _, err := PurchasedGrantActive(Active, nil, time.Now()); err == nil {
		t.Fatal("grant-bearing state accepted without window")
	}
	start := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	window, _ := billingperiod.New(start, start.Add(time.Hour))
	if _, err := PurchasedGrantActive(Paused, &window, start); err == nil {
		t.Fatal("non-grant state accepted a purchased window")
	}
	active, err := PurchasedGrantActive(Cancelled, nil, start)
	if err != nil || active {
		t.Fatalf("cancelled grant = %t, %v", active, err)
	}
}
