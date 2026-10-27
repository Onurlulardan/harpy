package application

import (
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestIssueGlobalBillingCommandIsExact(t *testing.T) {
	_, authority := platformBillingAuthorityFixture(
		t, "platform.billing.catalog.manage",
	)
	capability, err := IssueGlobalBillingCommand(
		authority, "billing.catalog.offering.create",
		"platform.billing.manage", 3, 2,
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = platformtx.NewGlobalCommandScopeRequest(capability); err != nil {
		t.Fatalf("issued capability is unusable: %v", err)
	}
	if _, err = IssueGlobalBillingCommand(
		authority, "billing.catalog.offering.create",
		"billing.manage", 3, 2,
	); err == nil {
		t.Fatal("tenant Billing permission minted a global capability")
	}
}
