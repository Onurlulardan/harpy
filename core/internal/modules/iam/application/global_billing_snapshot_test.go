package application

import (
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestGlobalBillingSnapshotIsReadOnlyAndActorBound(t *testing.T) {
	_, authority := platformBillingAuthorityFixture(
		t, "platform.billing.catalog.read",
	)
	capability, err := BuildGlobalBillingSnapshot(
		authority, "billing.catalog.offering.list",
		"platform.billing.read", 3, 2,
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = platformtx.NewGlobalSnapshotScopeRequest(capability); err != nil {
		t.Fatalf("issued snapshot capability is unusable: %v", err)
	}
	if _, err = BuildGlobalBillingSnapshot(
		authority, "billing.catalog.offering.list",
		"billing.read", 3, 2,
	); err == nil {
		t.Fatal("tenant permission minted global snapshot capability")
	}
}
