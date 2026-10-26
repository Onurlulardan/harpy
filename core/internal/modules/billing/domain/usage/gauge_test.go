package usage

import (
	"math"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
)

func TestGaugeAdjustmentPolicies(t *testing.T) {
	limit, _ := entitlement.IntegerValue(10)
	tests := []struct {
		name    string
		policy  entitlement.Enforcement
		current int64
		delta   int64
		allowed bool
		soft    bool
		after   int64
	}{
		{"increase", entitlement.Hard, 3, 2, true, false, 5},
		{"decrease", entitlement.Hard, 3, -2, true, false, 1},
		{"over-limit decrease", entitlement.Hard, 12, -1, true, false, 11},
		{"hard exceeded", entitlement.Hard, 9, 2, false, false, 11},
		{"soft exceeded", entitlement.Soft, 9, 2, true, true, 11},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := AdjustGauge(gaugeDefinition(t, test.policy, false),
				limit, test.current, test.delta)
			if err != nil || got.Allowed() != test.allowed ||
				got.SoftLimitExceeded() != test.soft || got.After() != test.after {
				t.Fatalf("got %+v, %v", got, err)
			}
		})
	}
}

func TestGaugeAdjustmentRejectsInvalidTotals(t *testing.T) {
	limit, _ := entitlement.IntegerValue(10)
	definition := gaugeDefinition(t, entitlement.Hard, false)
	for _, input := range [][2]int64{{0, 0}, {0, -1}, {math.MaxInt64, 1}} {
		if _, err := AdjustGauge(definition, limit, input[0], input[1]); err != ErrGauge {
			t.Fatalf("input %v: expected ErrGauge, got %v", input, err)
		}
	}
}

func gaugeDefinition(
	t *testing.T,
	policy entitlement.Enforcement,
	allowUnlimited bool,
) entitlement.Definition {
	t.Helper()
	key, _ := entitlement.NewKey("documents.bytes")
	definition, err := entitlement.NewDefinition(
		key, entitlement.Integer, entitlement.Gauge,
		entitlement.Never, policy, allowUnlimited,
	)
	if err != nil {
		t.Fatal(err)
	}
	return definition
}
