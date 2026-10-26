package discount

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
)

func TestApplyOrdersAndClampsDiscounts(t *testing.T) {
	net := testMoney(t, 1000, "USD")
	tenPercent, _ := money.NewBasisPoints(1000)
	rules := []Rule{
		mustBasisPointRule(t, testIdentifier(2), 20, tenPercent, ForeverDuration()),
		mustFixedRule(t, testIdentifier(3), 10, testMoney(t, 2000, "USD"), ForeverDuration()),
		mustBasisPointRule(t, testIdentifier(1), 20, tenPercent, ForeverDuration()),
	}
	result, err := Apply(net, 1, rules)
	if err != nil {
		t.Fatal(err)
	}
	if result.Net().MinorUnits() != 0 || result.Discount().MinorUnits() != 1000 {
		t.Fatalf("net=%d discount=%d", result.Net().MinorUnits(), result.Discount().MinorUnits())
	}
	applications := result.Applications()
	if len(applications) != 3 || applications[0].Identifier() != testIdentifier(3) ||
		applications[1].Identifier() != testIdentifier(1) ||
		applications[2].Identifier() != testIdentifier(2) {
		t.Fatalf("unexpected application order: %+v", applications)
	}
}

func TestApplyUsesRemainingNetAndHalfEven(t *testing.T) {
	net := testMoney(t, 15, "USD")
	fiftyPercent, _ := money.NewBasisPoints(5000)
	rules := []Rule{
		mustBasisPointRule(t, testIdentifier(1), 1, fiftyPercent, ForeverDuration()),
		mustBasisPointRule(t, testIdentifier(2), 2, fiftyPercent, ForeverDuration()),
	}
	result, err := Apply(net, 1, rules)
	if err != nil || result.Net().MinorUnits() != 3 || result.Discount().MinorUnits() != 12 {
		t.Fatalf("got net=%d discount=%d, %v", result.Net().MinorUnits(), result.Discount().MinorUnits(), err)
	}
}

func TestApplyHonorsDurationAndRejectsMismatch(t *testing.T) {
	net := testMoney(t, 100, "USD")
	once := mustFixedRule(t, testIdentifier(1), 1, testMoney(t, 20, "USD"), OnceDuration())
	result, err := Apply(net, 2, []Rule{once})
	if err != nil || result.Net().MinorUnits() != 100 || len(result.Applications()) != 0 {
		t.Fatalf("got %+v, %v", result, err)
	}
	mismatch := mustFixedRule(t, testIdentifier(2), 1, testMoney(t, 20, "EUR"), ForeverDuration())
	if _, err := Apply(net, 1, []Rule{mismatch}); err != ErrApply {
		t.Fatalf("expected ErrApply, got %v", err)
	}
}

func TestApplyRejectsDuplicateIdentity(t *testing.T) {
	rate, _ := money.NewBasisPoints(100)
	rule := mustBasisPointRule(t, testIdentifier(1), 1, rate, ForeverDuration())
	if _, err := Apply(testMoney(t, 100, "USD"), 1, []Rule{rule, rule}); err != ErrApply {
		t.Fatalf("expected ErrApply, got %v", err)
	}
}

func testMoney(t *testing.T, units int64, currency string) money.Money {
	t.Helper()
	value, err := money.New(units, currency, money.RegistryVersion)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func testIdentifier(last byte) Identifier {
	var identifier Identifier
	identifier[len(identifier)-1] = last
	return identifier
}

func mustBasisPointRule(t *testing.T, id Identifier, priority int32,
	rate money.BasisPoints, duration Duration) Rule {
	t.Helper()
	rule, err := NewBasisPointRule(id, priority, rate, duration)
	if err != nil {
		t.Fatal(err)
	}
	return rule
}

func mustFixedRule(t *testing.T, id Identifier, priority int32,
	fixed money.Money, duration Duration) Rule {
	t.Helper()
	rule, err := NewFixedRule(id, priority, fixed, duration)
	if err != nil {
		t.Fatal(err)
	}
	return rule
}
