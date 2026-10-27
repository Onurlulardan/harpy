package catalogquery

import (
	"strings"
	"testing"
)

func TestCatalogQueriesAreStaticQualifiedAndDeterministic(t *testing.T) {
	statements := map[string]string{
		"list": listOfferingsSQL, "get": getOfferingSQL,
		"campaign": getActiveCampaignSQL, "platform offering": getPlatformOfferingSQL,
		"platform discount": getPlatformDiscountSQL,
		"platform campaign": getPlatformCampaignSQL,
	}
	for name, statement := range statements {
		if statement == "" || strings.Contains(statement, "FROM billing_") ||
			strings.Contains(statement, "JOIN billing_") {
			t.Fatalf("%s statement lost qualified static SQL: %s", name, statement)
		}
	}
	if !strings.Contains(listOfferingsSQL, "selected.offering_key") ||
		!strings.Contains(listOfferingsSQL, "snapshot.offering_key > $1::text") ||
		!strings.Contains(listOfferingsSQL, "snapshot.version DESC") {
		t.Fatal("offering list lost keyset or deterministic version ordering")
	}
	for _, statement := range []string{
		listOfferingsSQL, getOfferingSQL, getPlatformOfferingSQL,
	} {
		if !strings.Contains(statement, "billing_catalog_offering_snapshot_read_v1") {
			t.Fatal("offering query bypasses the owner-owned snapshot read view")
		}
	}
	if !strings.Contains(getActiveCampaignSQL,
		"billing_catalog_campaign_snapshot_read_v1") ||
		!strings.Contains(getActiveCampaignSQL, "discount.priority") ||
		!strings.Contains(getActiveCampaignSQL, "discount.version_id") {
		t.Fatal("active campaign lost snapshot-safe deterministic selection")
	}
	for _, statement := range []string{
		getActiveCampaignSQL, getPlatformDiscountSQL, getPlatformCampaignSQL,
	} {
		if !strings.Contains(statement, "billing_promotion_discount_read_v2") ||
			!strings.Contains(statement, "'stable_offering_ids'") ||
			!strings.Contains(statement, "'legacy_offering_ids'") ||
			!strings.Contains(statement, "'used_count'") {
			t.Fatal("promotion query lost typed discount projection")
		}
	}
}
