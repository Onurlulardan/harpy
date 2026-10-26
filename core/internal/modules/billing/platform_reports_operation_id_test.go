package billing

import (
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestPlatformReportOperationIDIsStableUUIDv7(t *testing.T) {
	principalID := uuid.New()
	capability, err := platformtx.NewGlobalCommandCapabilityIssuer().Issue(
		principalID, operationSubscriptionTransition, "platform.billing.manage", 7, 5, 11, 13,
	)
	if err != nil {
		t.Fatal(err)
	}
	access, err := billingapi.NewGlobalAccessEvidence(
		principalID, 3, 5, 7, billingapi.PermissionPlatformBillingManage, capability,
	)
	if err != nil {
		t.Fatal(err)
	}
	subscriptionID, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	first, err := platformReportOperationID(access, subscriptionID, "request-1")
	second, secondErr := platformReportOperationID(access, subscriptionID, "request-1")
	if err != nil || secondErr != nil || first != second || first.Version() != 7 {
		t.Fatalf("operation ids = %s %s, errors = %v %v", first, second, err, secondErr)
	}
}
