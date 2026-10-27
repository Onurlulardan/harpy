package subscription

import (
	"strings"
	"testing"
)

func TestCurrentSubscriptionQueryPinsHistoricalOfferingVersionAndAuthority(t *testing.T) {
	required := []string{
		"billing_subscription_snapshot_read_v1",
		"candidate_count bigint",
		"offering_version_id uuid",
		"price_version_id uuid",
		"price_matches_offering boolean",
		"offering_matches boolean",
		"offering_document jsonb",
	}
	for _, fragment := range required {
		if !strings.Contains(getCurrentSQL, fragment) {
			t.Fatalf("current subscription SQL misses %q", fragment)
		}
	}
	for _, forbidden := range []string{
		"platform.billing_subscriptions",
		"platform.billing_offerings",
		"platform.billing_price_versions",
	} {
		if strings.Contains(getCurrentSQL, forbidden) {
			t.Fatalf("historical subscription SQL contains visibility filter %q", forbidden)
		}
	}
}
