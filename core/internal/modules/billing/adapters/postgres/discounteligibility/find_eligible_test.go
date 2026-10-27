package discounteligibility

import (
	"strings"
	"testing"
	"time"

	discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
	"github.com/google/uuid"
)

func TestEligibilityRecordBuildsRestrictedTypedPolicy(t *testing.T) {
	now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	basisPoints := int32(1500)
	offeringID := mustAdapterV7(t)
	record := eligibilityRecord{
		candidateCount: 1, subscriptionCurrency: "USD", code: "launch",
		kind: "basis_point", basisPoints: &basisPoints, duration: "once",
		priority: 2, expectedOfferingCount: 1, validStart: now,
		available: true, offeringAllowed: true,
	}
	policy, err := record.policy([]discountdomain.Identifier{
		discountdomain.Identifier(offeringID),
	})
	code, _ := discountdomain.NormalizeCode("launch")
	if err != nil || policy.AllOfferings() || len(policy.OfferingIDs()) != 1 ||
		policy.Code() != code {
		t.Fatalf("policy=%+v error=%v", policy, err)
	}
	record.alreadyRedeemed = true
	if record.eligible() {
		t.Fatal("an existing subscription redemption remained eligible")
	}
	record.alreadyRedeemed, record.available = false, false
	if record.eligible() {
		t.Fatal("an unavailable latest discount version remained eligible")
	}
}

func TestEligibilitySQLPinsTenantAndCurrentGlobalAuthorities(t *testing.T) {
	expected := []struct {
		document string
		marker   string
	}{
		{findEligibleSQL, "billing_discount_eligibility_snapshot_read_v1"},
		{findEligibleSQL, "$1::uuid, $2::uuid"},
		{findEligibleSQL, "$5::timestamptz, $6::uuid"},
		{resolveAuthoritySQL, "billing_resolve_current_tenant_authority_v1"},
		{resolveAuthoritySQL, "billing_resolve_current_global_authority_v1"},
		{listOfferingsSQL, "billing_discount_offerings_snapshot_read_v1"},
	}
	for _, item := range expected {
		if !strings.Contains(item.document, item.marker) {
			t.Fatalf("SQL missing authority marker %q", item.marker)
		}
	}
}

func mustAdapterV7(t *testing.T) uuid.UUID {
	t.Helper()
	value, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return value
}
