package discount

import (
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
)

func TestPolicyRequiresExplicitApplicability(t *testing.T) {
	code, _ := NormalizeCode("SAVE10")
	rate, _ := money.NewBasisPoints(1000)
	now := policyInstant()
	id := Identifier{15: 1}
	if _, err := NewBasisPointPolicy(code, rate, OnceDuration(), 1, nil, nil,
		now, nil, true, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := NewBasisPointPolicy(code, rate, OnceDuration(), 1, nil, nil,
		now, nil, false, []Identifier{id}); err != nil {
		t.Fatal(err)
	}
	invalid := [][]Identifier{nil, {id, id}}
	for _, identifiers := range invalid {
		if _, err := NewBasisPointPolicy(code, rate, OnceDuration(), 1, nil, nil,
			now, nil, false, identifiers); err == nil {
			t.Fatal("invalid restricted applicability accepted")
		}
	}
	if _, err := NewBasisPointPolicy(code, rate, OnceDuration(), 1, nil, nil,
		now, nil, true, []Identifier{id}); err == nil {
		t.Fatal("all plus explicit list accepted")
	}
}

func TestPolicyCopiesLimitsIdentifiersAndBuildsRule(t *testing.T) {
	code, _ := NormalizeCode("FIXED")
	amount, _ := money.New(500, "USD", money.RegistryVersion)
	limit := int64(10)
	identifiers := []Identifier{{15: 2}, {15: 1}}
	policy, err := NewFixedPolicy(code, amount, ForeverDuration(), 3, &limit, &limit,
		policyInstant(), nil, false, identifiers)
	if err != nil {
		t.Fatal(err)
	}
	limit = 99
	identifiers[0] = Identifier{}
	if *policy.GlobalLimit() != 10 || policy.OfferingIDs()[0][15] != 1 {
		t.Fatal("caller mutation changed policy")
	}
	rule, err := policy.Rule(Identifier{15: 9})
	fixed, ok := rule.FixedAmount()
	if err != nil || !ok || fixed.MinorUnits() != 500 || rule.Priority() != 3 {
		t.Fatalf("rule=%#v fixed=%#v ok=%t err=%v", rule, fixed, ok, err)
	}
}

func TestPolicyRejectsInvalidLimitsAndInterval(t *testing.T) {
	code, _ := NormalizeCode("SAVE")
	rate, _ := money.NewBasisPoints(100)
	zero := int64(0)
	now := policyInstant()
	if _, err := NewBasisPointPolicy(code, rate, OnceDuration(), 0, &zero, nil,
		now, nil, true, nil); err == nil {
		t.Fatal("zero limit accepted")
	}
	if _, err := NewBasisPointPolicy(code, rate, OnceDuration(), 0, nil, nil,
		now, &now, true, nil); err == nil {
		t.Fatal("empty interval accepted")
	}
}

func policyInstant() time.Time {
	return time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
}
