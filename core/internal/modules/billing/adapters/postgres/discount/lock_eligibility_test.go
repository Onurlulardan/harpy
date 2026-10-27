package discount

import (
	"context"
	"crypto/sha256"
	"strings"
	"testing"
	"time"

	discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
)

func TestLockEligibilityMapsPinnedPolicy(t *testing.T) {
	authorityID := testV7()
	discountVersionID, offeringID, priceID := testV7(), testV7(), testV7()
	now := time.Date(2026, 7, 17, 6, 0, 0, 0, time.UTC)
	digest := sha256.Sum256([]byte("policy"))
	globalLimit, tenantLimit := int64(20), int64(2)
	storage, executions := testStore(
		scriptedRow{values: []any{authorityID}},
		scriptedRow{values: []any{
			discountVersionID, "LAUNCH", offeringID, priceID, "USD",
			now, now.Add(time.Hour), &globalLimit, int64(3), &tenantLimit,
			int64(1), true, digest[:],
		}},
	)
	code, _ := discountdomain.NormalizeCode("launch")
	result, err := storage.lockEligibility(
		context.Background(), code, testV7(), offeringID, priceID, now,
	)
	if err != nil || result.DiscountVersionID != discountVersionID ||
		result.PolicySHA256 != digest || !result.OfferingAllowed || *executions != 2 {
		t.Fatalf("result=%+v executions=%d err=%v", result, *executions, err)
	}
}

func TestLockEligibilityPinsClockAndCurrentGlobalAuthority(t *testing.T) {
	for _, marker := range []string{
		"version_row.valid_start <= $7",
		"discount_row.authority_id",
		"billing_global_authorities",
		"global_group.state = 'v2'",
		"global_domain.writer = 'v2'",
		"existing.subscription_id = $4",
		"candidate.available",
	} {
		if !strings.Contains(lockEligibilitySQL, marker) {
			t.Fatalf("lock eligibility SQL missing %q", marker)
		}
	}
}
