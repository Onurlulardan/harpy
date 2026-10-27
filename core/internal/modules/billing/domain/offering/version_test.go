package offering

import (
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
)

func TestVersionAcceptsTypedRegistryGrantsAndDefensiveCopies(t *testing.T) {
	now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	amount, _ := money.New(9900, "USD", money.RegistryVersion)
	price, _ := NewPrice(Monthly, amount)
	entry := entitlement.RegistryV1ByKey()["iam.active_users"]
	limit, _ := entitlement.IntegerValue(500)
	grant, _ := NewGrant(entry.Definition, limit)
	prices, grants := []Price{price}, []Grant{grant}
	version, err := NewVersion("Pro", true, true, false, 14, 3, now, nil, prices, grants)
	if err != nil {
		t.Fatal(err)
	}
	prices[0] = Price{}
	grants[0] = Grant{}
	if version.Prices()[0].Cadence() != Monthly ||
		version.Grants()[0].Definition().Key() != "iam.active_users" {
		t.Fatal("caller mutation changed version")
	}
	copyPrices := version.Prices()
	copyPrices[0] = Price{}
	if version.Prices()[0].Cadence() != Monthly {
		t.Fatal("projection mutation changed version")
	}
}

func TestVersionRejectsDuplicatePriceAndGrant(t *testing.T) {
	now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	amount, _ := money.New(0, "USD", money.RegistryVersion)
	price, _ := NewPrice(Monthly, amount)
	entry := entitlement.RegistryV1ByKey()["iam.active_users"]
	grant, _ := NewGrant(entry.Definition, entitlement.UnlimitedValue())
	if _, err := NewVersion("Base", true, true, true, 0, 0, now, nil,
		[]Price{price, price}, []Grant{grant}); err == nil {
		t.Fatal("duplicate price accepted")
	}
	if _, err := NewVersion("Base", true, true, true, 0, 0, now, nil,
		[]Price{price}, []Grant{grant, grant}); err == nil {
		t.Fatal("duplicate grant accepted")
	}
}

func TestVersionRejectsInvalidIntervalAndNegativePrice(t *testing.T) {
	now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	amount, _ := money.New(-1, "USD", money.RegistryVersion)
	if _, err := NewPrice(Monthly, amount); err == nil {
		t.Fatal("negative price accepted")
	}
	zero, _ := money.New(0, "USD", money.RegistryVersion)
	price, _ := NewPrice(Monthly, zero)
	end := now
	if _, err := NewVersion("Base", true, true, true, 0, 0, now, &end,
		[]Price{price}, nil); err == nil {
		t.Fatal("empty interval accepted")
	}
}

func TestVersionPreservesImmutableAvailability(t *testing.T) {
	now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	amount, _ := money.New(0, "USD", money.RegistryVersion)
	price, _ := NewPrice(Monthly, amount)
	version, err := NewVersion("Hidden", true, false, false, 0, 0, now, nil,
		[]Price{price}, nil)
	if err != nil || version.Available() {
		t.Fatalf("availability = %t, %v", version.Available(), err)
	}
}

func TestOfferingKeyVocabulary(t *testing.T) {
	for _, value := range []string{"base-plan", "pro.monthly", "team_2"} {
		if _, err := NewKey(value); err != nil {
			t.Fatalf("valid key %q: %v", value, err)
		}
	}
	for _, value := range []string{"a", "Base", "-base", "base."} {
		if _, err := NewKey(value); err == nil {
			t.Fatalf("invalid key %q accepted", value)
		}
	}
}
