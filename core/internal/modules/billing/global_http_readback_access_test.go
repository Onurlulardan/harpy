package billing

import (
	"errors"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestCatalogReadbackUsesExactManageCommandEvidence(t *testing.T) {
	t.Parallel()
	principalID := uuid.New()
	command, err := platformtx.NewGlobalCommandAccessIssuer().Issue(
		principalID, discountPublishOperation,
		string(billingapi.PermissionPlatformBillingManage), 7, 5, 3, 2,
	)
	if err != nil {
		t.Fatal(err)
	}
	access, err := billingapi.NewGlobalCommandAccessEvidence(
		principalID, 11, 5, 7,
		billingapi.PermissionPlatformBillingManage, command,
	)
	if err != nil {
		t.Fatal(err)
	}
	capability, err := catalogReadbackCapability(access, offeringReadOperation)
	if err != nil || !capability.Allows(
		principalID, discountPublishOperation,
		string(billingapi.PermissionPlatformBillingManage), 7, 5,
	) {
		t.Fatalf("manage readback rejected: %v", err)
	}

	reportCommand, err := platformtx.NewGlobalCommandAccessIssuer().Issue(
		principalID, operationSubscriptionTransition,
		string(billingapi.PermissionPlatformBillingManage), 7, 5, 3, 2,
	)
	if err != nil {
		t.Fatal(err)
	}
	reportAccess, err := billingapi.NewGlobalCommandAccessEvidence(
		principalID, 11, 5, 7,
		billingapi.PermissionPlatformBillingManage, reportCommand,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = catalogReadbackCapability(reportAccess, offeringReadOperation)
	if !errors.Is(err, billingapi.ErrForbidden) {
		t.Fatalf("unrelated command readback error = %v", err)
	}
}
