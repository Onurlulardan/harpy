package transaction

import "testing"

func TestPublicGlobalSnapshotCapabilityHasExactReadAllowlist(t *testing.T) {
	issuer := NewPublicGlobalSnapshotCapabilityIssuer()
	for _, operation := range []PublicGlobalOperation{
		PublicBillingCatalogGet, PublicBillingCatalogList,
	} {
		capability, err := issuer.Issue(operation, 5, 3)
		if err != nil || !capability.valid() {
			t.Fatalf("public operation %q rejected: %v", operation, err)
		}
	}
	for _, operation := range []PublicGlobalOperation{
		"", "billing.catalog.create", "billing.subscription.get",
	} {
		if _, err := issuer.Issue(operation, 5, 3); err == nil {
			t.Fatalf("public operation %q was accepted", operation)
		}
	}
	if _, err := issuer.Issue(PublicBillingCatalogGet, 0, 3); err == nil {
		t.Fatal("public capability without group generation was accepted")
	}
}
