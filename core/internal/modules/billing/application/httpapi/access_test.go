package httpapi

import (
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestGlobalAccessEvidencePreservesReadOnlyCapabilityKind(t *testing.T) {
	principalID := uuid.New()
	capability, _ := platformtx.NewGlobalSnapshotCapabilityIssuer().Issue(
		principalID, "billing.catalog.offering.list", "platform.billing.read",
		7, 5, 3, 2,
	)
	access, err := NewGlobalSnapshotAccessEvidence(
		principalID, 11, 5, 7, PermissionPlatformBillingRead, capability,
	)
	if err != nil || !access.ValidFor(PermissionPlatformBillingRead) {
		t.Fatalf("global snapshot access = %#v, %v", access, err)
	}
	if _, ok := access.GlobalSnapshotCapability(); !ok {
		t.Fatal("global snapshot evidence was not retained")
	}
	if _, ok := access.GlobalCommandCapability(); ok {
		t.Fatal("global snapshot evidence widened to command")
	}
}

func TestGlobalAccessEvidenceRequiresExactOperation(t *testing.T) {
	principalID := uuid.New()
	command, err := platformtx.NewGlobalCommandAccessIssuer().Issue(
		principalID, "billing.catalog.offering.create", "platform.billing.manage",
		7, 5, 3, 2,
	)
	if err != nil {
		t.Fatal(err)
	}
	commandAccess, err := NewGlobalCommandAccessEvidence(
		principalID, 11, 5, 7, PermissionPlatformBillingManage, command,
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, valid := commandAccess.GlobalCommandCapabilitiesFor(
		"billing.catalog.offering.create",
	); !valid {
		t.Fatal("exact command operation was rejected")
	}
	if _, _, valid := commandAccess.GlobalCommandCapabilitiesFor(
		"billing.catalog.offering.retire",
	); valid {
		t.Fatal("command evidence widened to another operation")
	}
	snapshot, err := platformtx.NewGlobalSnapshotCapabilityIssuer().Issue(
		principalID, "billing.catalog.offering.list", "platform.billing.read",
		7, 5, 3, 2,
	)
	if err != nil {
		t.Fatal(err)
	}
	snapshotAccess, err := NewGlobalSnapshotAccessEvidence(
		principalID, 11, 5, 7, PermissionPlatformBillingRead, snapshot,
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, valid := snapshotAccess.GlobalSnapshotCapabilityFor(
		"billing.catalog.discount.list",
	); valid {
		t.Fatal("snapshot evidence widened to another operation")
	}
}
