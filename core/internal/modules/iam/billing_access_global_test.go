package iam

import (
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestBillingGlobalPermissionsSeparateSnapshotAndCommandEvidence(t *testing.T) {
	read, readOK := resolveBillingPermission(
		"platform.billing.read", "billing.catalog.offering.list",
	)
	manage, manageOK := resolveBillingPermission(
		"platform.billing.manage", "billing.catalog.offering.create",
	)
	if !readOK || !read.global || read.command || !manageOK || !manage.command {
		t.Fatalf("permission modes read=%#v manage=%#v", read, manage)
	}
	principalID, tenantID := uuid.New(), uuid.New()
	snapshot, _ := platformtx.NewGlobalSnapshotCapabilityIssuer().Issue(
		principalID, "billing.catalog.offering.list", "platform.billing.read",
		7, 5, 3, 2,
	)
	snapshotDecision, err := NewGlobalBillingSnapshotAccessDecision(
		principalID, tenantID, 11, 5, 7, snapshot,
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := snapshotDecision.GlobalSnapshotCapability(); !ok {
		t.Fatal("snapshot decision lost snapshot capability")
	}
	if _, ok := snapshotDecision.GlobalCommandAccess(); ok {
		t.Fatal("snapshot decision exposed command capability")
	}
}
