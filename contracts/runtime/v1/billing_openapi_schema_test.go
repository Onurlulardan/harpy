package runtimev1

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestBillingOpenAPIDocumentsAreStrictJSONAndResolveLocalFiles(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	files := []string{
		"core-api-v1.components.json",
		"billing-common-v2.schemas.json", "billing-catalog-v2.schemas.json",
		"billing-subscription-v2.schemas.json", "billing-v1-compat.schemas.json",
		"billing-public-tenant-v2.paths.json", "billing-platform-offerings-v2.paths.json",
		"billing-platform-discounts-v2.paths.json", "billing-platform-campaigns-v2.paths.json",
		"billing-platform-reports-v2.paths.json", "billing-webhook-v2.paths.json",
		"billing-v1-public-tenant.paths.json", "billing-v1-admin-core.paths.json",
		"billing-v1-admin-promotions.paths.json",
	}
	for _, name := range files {
		document := readOpenAPIDocument(t, filepath.Join(directory, name))
		assertLocalReferences(t, directory, document)
	}
}

func TestBillingCatalogSchemasPinCASIntervalsLimitsAndApplicability(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	common := readOpenAPIDocument(t, filepath.Join(directory, "billing-common-v2.schemas.json"))
	commonSchemas := common["schemas"].(map[string]any)
	revision := commonSchemas["PositiveRevision"].(map[string]any)
	if revision["minimum"] != float64(1) || revision["format"] != "int64" {
		t.Fatalf("positive revision schema=%v", revision)
	}
	catalog := readOpenAPIDocument(t, filepath.Join(directory, "billing-catalog-v2.schemas.json"))
	schemas := catalog["schemas"].(map[string]any)
	for _, name := range []string{
		"OfferingVersionMutation", "DiscountVersionMutation", "CampaignVersionMutation", "Retire",
	} {
		schema := schemas[name].(map[string]any)
		if !stringListContains(schema["required"].([]any), "expected_revision") {
			t.Fatalf("schema %s does not require expected_revision", name)
		}
	}
	offering := schemas["OfferingVersionInput"].(map[string]any)
	offeringProperties := offering["properties"].(map[string]any)
	if offeringProperties["effective_start"] == nil || offeringProperties["effective_end"] == nil ||
		offering["x-harpy-interval"] != "half-open [effective_start,effective_end)" {
		t.Fatalf("offering effective interval=%v", offering)
	}
	discount := schemas["DiscountVersionInput"].(map[string]any)
	discountProperties := discount["properties"].(map[string]any)
	for _, name := range []string{
		"all_offerings", "offering_version_ids", "valid_start", "valid_end",
		"global_use_limit", "per_tenant_use_limit",
	} {
		if discountProperties[name] == nil {
			t.Fatalf("discount property %s is missing", name)
		}
	}
	if len(discount["oneOf"].([]any)) != 2 ||
		discountProperties["global_use_limit"].(map[string]any)["minimum"] != float64(1) ||
		discountProperties["per_tenant_use_limit"].(map[string]any)["minimum"] != float64(1) {
		t.Fatalf("discount applicability or limits=%v", discount)
	}
	identifiers := discountProperties["offering_version_ids"].(map[string]any)
	pattern := identifiers["items"].(map[string]any)["pattern"].(string)
	if !strings.Contains(pattern, "-7") || identifiers["uniqueItems"] != true {
		t.Fatalf("offering version IDs are not unique UUIDv7: %v", identifiers)
	}
	campaign := schemas["CampaignVersionInput"].(map[string]any)
	campaignIDs := campaign["properties"].(map[string]any)["discount_version_ids"].(map[string]any)
	if campaignIDs["maxItems"] != float64(256) || campaignIDs["minItems"] != nil {
		t.Fatalf("banner-only campaign contract mismatch: %v", campaignIDs)
	}
}

func TestBillingUsageLimitEnvelopeHasExactLosslessDetails(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	document := readOpenAPIDocument(t, filepath.Join(directory, "billing-common-v2.schemas.json"))
	schemas := document["schemas"].(map[string]any)
	details := schemas["UsageLimitDetails"].(map[string]any)
	properties := details["properties"].(map[string]any)
	want := []string{
		"entitlement_key", "current", "requested", "limit",
		"period_start", "period_end", "reset_policy",
	}
	if len(properties) != len(want) || len(details["required"].([]any)) != len(want) {
		t.Fatalf("usage details members=%v", properties)
	}
	for _, name := range want {
		if properties[name] == nil || !stringListContains(details["required"].([]any), name) {
			t.Fatalf("usage detail %s missing", name)
		}
	}
	for _, name := range []string{"current", "requested", "limit"} {
		if properties[name].(map[string]any)["type"] != "string" {
			t.Fatalf("usage counter %s is not lossless string", name)
		}
	}
}

func TestBillingCatalogMutationPathsDeclareStaleRevisionConflict(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	core := readOpenAPIDocument(t, filepath.Join(directory, "core-api-v1.openapi.yaml"))
	paths := core["paths"].(map[string]any)
	for _, path := range []string{
		"/api/platform/billing/offerings/{offering_id}/versions",
		"/api/platform/billing/offerings/{offering_id}/retire",
		"/api/platform/billing/discounts/{discount_id}/versions",
		"/api/platform/billing/discounts/{discount_id}/retire",
		"/api/platform/billing/campaigns/{campaign_id}/versions",
		"/api/platform/billing/campaigns/{campaign_id}/retire",
	} {
		operation := resolveCorePathItem(t, directory, paths, path)["post"].(map[string]any)
		responses := operation["responses"].(map[string]any)
		conflict := responses["409"].(map[string]any)["$ref"].(string)
		if operation["x-harpy-cas"] != "expected_revision" ||
			!strings.HasSuffix(conflict, "/RevisionStale") {
			t.Fatalf("CAS path %s=%v", path, operation)
		}
	}
}

func stringListContains(values []any, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
