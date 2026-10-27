package usage

import (
	"math"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
)

func TestCounterAdmissionPolicies(t *testing.T) {
	hard := counterDefinition(t, entitlement.Hard, true)
	soft := counterDefinition(t, entitlement.Soft, false)
	limit, err := entitlement.IntegerValue(10)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name      string
		def       entitlement.Definition
		committed int64
		reserved  int64
		requested int64
		allowed   bool
		soft      bool
	}{
		{"below", hard, 3, 2, 4, true, false},
		{"boundary", hard, 3, 2, 5, true, false},
		{"hard exceeded", hard, 3, 2, 6, false, false},
		{"soft exceeded", soft, 3, 2, 6, true, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Admit(test.def, limit, test.committed, test.reserved, test.requested)
			if err != nil || got.Allowed() != test.allowed ||
				got.SoftLimitExceeded() != test.soft {
				t.Fatalf("got %+v, %v", got, err)
			}
		})
	}
}

func TestCounterAdmissionAllowsTypedUnlimited(t *testing.T) {
	decision, err := Admit(
		counterDefinition(t, entitlement.Hard, true),
		entitlement.UnlimitedValue(), math.MaxInt64-1, 0, 1,
	)
	if err != nil || !decision.Allowed() || decision.Projected() != math.MaxInt64 {
		t.Fatalf("got %+v, %v", decision, err)
	}
}

func TestCounterAdmissionRejectsInvalidAndOverflow(t *testing.T) {
	limit, _ := entitlement.IntegerValue(10)
	definition := counterDefinition(t, entitlement.Hard, false)
	for _, input := range [][3]int64{{-1, 0, 1}, {0, -1, 1}, {0, 0, 0}, {math.MaxInt64, 0, 1}} {
		if _, err := Admit(definition, limit, input[0], input[1], input[2]); err != ErrAdmission {
			t.Fatalf("input %v: expected ErrAdmission, got %v", input, err)
		}
	}
	if _, err := Admit(definition, entitlement.UnlimitedValue(), 0, 0, 1); err != ErrAdmission {
		t.Fatalf("unallowed unlimited: expected ErrAdmission, got %v", err)
	}
}

func counterDefinition(
	t *testing.T,
	enforcement entitlement.Enforcement,
	allowUnlimited bool,
) entitlement.Definition {
	t.Helper()
	key, err := entitlement.NewKey("iam.active_users")
	if err != nil {
		t.Fatal(err)
	}
	definition, err := entitlement.NewDefinition(
		key, entitlement.Integer, entitlement.Counter,
		entitlement.BillingPeriod, enforcement, allowUnlimited,
	)
	if err != nil {
		t.Fatal(err)
	}
	return definition
}
