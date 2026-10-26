package transaction

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

func TestProviderTargetMintsOnlyBillingAuthorityForMappedTenant(t *testing.T) {
	lookup, _ := NewProviderLookup(
		"testpay", "acct_10", "subscription", "sub_20", "evt_30", [32]byte{1},
	)
	evidence := providerEvidenceFixture(lookup)
	target, err := issueProviderTarget(lookup, evidence)
	if err != nil {
		t.Fatal(err)
	}
	state := &providerScopeState{
		lookup: lookup, target: target, transaction: &postgres.RuntimeTransaction{},
	}
	state.active.Store(true)
	scope := ProviderTenantScope{state: state}
	authority, err := MintProviderOwnerAuthority(scope, "billing")
	if err != nil || !authority.OwnsProviderScope(scope) ||
		authority.ScopeID() != evidence.TenantID || authority.GroupID() != "billing_forms" ||
		authority.GroupGeneration() != 7 || authority.Domain() != "billing" ||
		authority.Generation() != 11 {
		t.Fatalf("authority=%+v error=%v", authority, err)
	}
	if _, err := MintProviderOwnerAuthority(scope, "forms"); err != ErrScopeOwner {
		t.Fatalf("forms owner error=%v", err)
	}
	state.active.Store(false)
	if authority.OwnsProviderScope(scope) || authority.ScopeID() != uuid.Nil {
		t.Fatal("provider authority escaped closed scope")
	}
}

func TestProviderTargetRejectsLookupOrMappingDrift(t *testing.T) {
	lookup, _ := NewProviderLookup(
		"testpay", "acct_10", "subscription", "sub_20", "evt_30", [32]byte{1},
	)
	tests := []ProviderTargetEvidence{
		func() ProviderTargetEvidence {
			value := providerEvidenceFixture(lookup)
			value.LookupBindingSHA256[0] ^= 1
			return value
		}(),
		func() ProviderTargetEvidence {
			value := providerEvidenceFixture(lookup)
			value.SubscriptionID = uuid.Nil
			return value
		}(),
		func() ProviderTargetEvidence {
			value := providerEvidenceFixture(lookup)
			value.FormsGeneration = 0
			return value
		}(),
	}
	for _, evidence := range tests {
		if _, err := issueProviderTarget(lookup, evidence); err == nil {
			t.Fatalf("accepted invalid evidence: %+v", evidence)
		}
	}
}

func providerEvidenceFixture(lookup ProviderLookup) ProviderTargetEvidence {
	return ProviderTargetEvidence{
		TenantID:          uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		MappingID:         uuid.MustParse("22222222-2222-4222-8222-222222222222"),
		ProviderAccountID: uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		SubscriptionID:    uuid.MustParse("44444444-4444-4444-8444-444444444444"),
		MappingRevision:   3, GroupGeneration: 7,
		BillingGeneration: 11, FormsGeneration: 13,
		LookupBindingSHA256: ProviderLookupBinding(lookup),
	}
}
