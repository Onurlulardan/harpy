package application

import (
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestGlobalBillingCommandAccessCarriesSameGenerationReadback(t *testing.T) {
	_, authority := platformBillingAuthorityFixture(
		t, "platform.billing.coupons.manage",
	)
	access, err := BuildGlobalBillingCommandAccess(
		authority, "billing.catalog.discount.create", "platform.billing.manage",
		3, 2,
	)
	if err != nil {
		t.Fatal(err)
	}
	command, readback, ok := access.Capabilities()
	if !ok {
		t.Fatal("command access is unusable")
	}
	if _, err = platformtx.NewGlobalCommandScopeRequest(command); err != nil {
		t.Fatalf("command proof: %v", err)
	}
	if _, err = platformtx.NewGlobalSnapshotScopeRequest(readback); err != nil {
		t.Fatalf("readback proof: %v", err)
	}
}
