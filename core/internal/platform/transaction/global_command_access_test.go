package transaction

import (
	"testing"

	"github.com/google/uuid"
)

func TestGlobalCommandAccessBindsCommandAndReadbackEvidence(t *testing.T) {
	access, err := NewGlobalCommandAccessIssuer().Issue(
		uuid.New(), "billing.catalog.offering.create", "platform.billing.manage",
		7, 5, 3, 2,
	)
	if err != nil {
		t.Fatal(err)
	}
	command, snapshot, ok := access.Capabilities()
	if !ok {
		t.Fatal("issued access is unusable")
	}
	if _, err = NewGlobalCommandScopeRequest(command); err != nil {
		t.Fatalf("command capability: %v", err)
	}
	if _, err = NewGlobalSnapshotScopeRequest(snapshot); err != nil {
		t.Fatalf("readback capability: %v", err)
	}
}

func TestGlobalCommandAccessRejectsMixedOrZeroEvidence(t *testing.T) {
	actorID := uuid.New()
	command, _ := NewGlobalCommandCapabilityIssuer().Issue(
		actorID, "billing.catalog.offering.create", "platform.billing.manage",
		7, 5, 3, 2,
	)
	snapshot, _ := NewGlobalSnapshotCapabilityIssuer().Issue(
		actorID, "billing.catalog.offering.retire", "platform.billing.manage",
		7, 5, 3, 2,
	)
	mixed := GlobalCommandAccess{command: command, readback: snapshot, issued: true}
	if _, _, ok := mixed.Capabilities(); ok {
		t.Fatal("mixed operation evidence was accepted")
	}
	if _, _, ok := (GlobalCommandAccess{}).Capabilities(); ok {
		t.Fatal("zero evidence was accepted")
	}
}
