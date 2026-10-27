package transaction

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestGlobalScopeRequestSeparatesCommandAndPublicSnapshot(t *testing.T) {
	actorID := uuid.New()
	command, err := NewGlobalCommandCapabilityIssuer().Issue(
		actorID, "billing.catalog.publish", "platform.billing.catalog.manage",
		2, 3, 5, 7,
	)
	if err != nil {
		t.Fatal(err)
	}
	commandRequest, err := NewGlobalCommandScopeRequest(command)
	if err != nil || commandRequest.mode() != Command ||
		commandRequest.actorID() != actorID {
		t.Fatalf("command request = %#v, %v", commandRequest, err)
	}
	snapshot, _ := NewPublicGlobalSnapshotCapabilityIssuer().Issue(
		PublicBillingCatalogList, 5, 7,
	)
	snapshotRequest, err := NewPublicGlobalSnapshotScopeRequest(snapshot)
	if err != nil || snapshotRequest.mode() != SnapshotQuery ||
		snapshotRequest.actorID() != uuid.Nil {
		t.Fatalf("public snapshot request = %#v, %v", snapshotRequest, err)
	}
	if _, err := NewGlobalCommandScopeRequest(GlobalCommandCapability{}); !errors.Is(
		err, ErrInvalidGlobalScopeRequest,
	) {
		t.Fatalf("empty command request error = %v", err)
	}
	if _, err := NewPublicGlobalSnapshotScopeRequest(
		PublicGlobalSnapshotCapability{},
	); !errors.Is(err, ErrInvalidGlobalScopeRequest) {
		t.Fatalf("empty snapshot request error = %v", err)
	}
}

func TestGlobalScopeRequestAcceptsAuthenticatedSnapshotOnlyAsRead(t *testing.T) {
	capability, err := NewGlobalSnapshotCapabilityIssuer().Issue(
		uuid.New(), "billing.catalog.offering.list", "platform.billing.read",
		7, 5, 3, 2,
	)
	if err != nil {
		t.Fatal(err)
	}
	request, err := NewGlobalSnapshotScopeRequest(capability)
	if err != nil || request.mode() != SnapshotQuery || request.actorID() == uuid.Nil ||
		request.operationID() != "billing.catalog.offering.list" {
		t.Fatalf("authenticated snapshot request = %#v, %v", request, err)
	}
}
