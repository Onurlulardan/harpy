package transaction

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestGlobalSnapshotAuthorityCapturesGenerationFence(t *testing.T) {
	state := &globalScopeState{
		environmentID: uuid.New(), mode: SnapshotQuery,
		groupGeneration: 3, billingGeneration: 7,
	}
	state.active.Store(true)
	scope := GlobalScope{state: state}
	authority, err := MintGlobalSnapshotAuthority(scope, "billing")
	if err != nil || !authority.Owns(scope) || authority.EnvironmentID() != state.environmentID ||
		authority.GroupGeneration() != 3 || authority.BillingGeneration() != 7 {
		t.Fatalf("snapshot authority = %#v, %v", authority, err)
	}
	state.active.Store(false)
	if authority.Owns(scope) || authority.GroupGeneration() != 0 {
		t.Fatal("snapshot authority survived scope closure")
	}
}

func TestGlobalSnapshotAuthorityRejectsWrongModeOwnerAndFence(t *testing.T) {
	state := &globalScopeState{
		environmentID: uuid.New(), mode: Command,
		groupGeneration: 1, billingGeneration: 1,
	}
	state.active.Store(true)
	scope := GlobalScope{state: state}
	if _, err := MintGlobalSnapshotAuthority(scope, "billing"); !errors.Is(err, ErrScopeOwner) {
		t.Fatalf("command error = %v", err)
	}
	state.mode = SnapshotQuery
	if _, err := MintGlobalSnapshotAuthority(scope, "forms"); !errors.Is(err, ErrScopeOwner) {
		t.Fatalf("owner error = %v", err)
	}
	state.billingGeneration = 0
	if _, err := MintGlobalSnapshotAuthority(scope, "billing"); !errors.Is(err, ErrScopeOwner) {
		t.Fatalf("fence error = %v", err)
	}
}
