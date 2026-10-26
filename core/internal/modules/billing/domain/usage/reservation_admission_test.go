package usage

import (
	"math"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
)

func TestGaugeReservationIncludesCommittedAndPendingBytes(t *testing.T) {
	key, _ := entitlement.NewKey("documents.storage_bytes")
	definition, err := entitlement.NewDefinition(
		key, entitlement.Integer, entitlement.Gauge,
		entitlement.Never, entitlement.Hard, true,
	)
	if err != nil {
		t.Fatal(err)
	}
	limit, _ := entitlement.IntegerValue(10)
	tests := []struct {
		name                    string
		gauge, reserved, amount int64
		allowed                 bool
	}{
		{"below", 3, 2, 4, true},
		{"boundary", 3, 2, 5, true},
		{"exceeded", 3, 2, 6, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, callErr := AdmitReservation(
				definition, limit, 0, test.reserved, test.gauge, test.amount)
			if callErr != nil || result.Allowed() != test.allowed {
				t.Fatalf("result=%+v err=%v", result, callErr)
			}
		})
	}
}

func TestGaugeReservationRejectsCounterStateAndOverflow(t *testing.T) {
	key, _ := entitlement.NewKey("documents.storage_bytes")
	definition, _ := entitlement.NewDefinition(
		key, entitlement.Integer, entitlement.Gauge,
		entitlement.Never, entitlement.Hard, true,
	)
	limit, _ := entitlement.IntegerValue(math.MaxInt64)
	for _, values := range [][4]int64{
		{1, 0, 0, 1}, {0, -1, 0, 1}, {0, 0, -1, 1},
		{0, 1, math.MaxInt64, 1},
	} {
		if _, err := AdmitReservation(
			definition, limit, values[0], values[1], values[2], values[3],
		); err != ErrAdmission {
			t.Fatalf("values=%v err=%v", values, err)
		}
	}
}
