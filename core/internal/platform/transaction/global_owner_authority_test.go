package transaction

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestMintGlobalOwnerAuthorityProjectsExactBillingScope(t *testing.T) {
	environmentID := uuid.New()
	state := &globalScopeState{
		environmentID: environmentID, mode: Command,
		groupGeneration: 5, billingGeneration: 7,
	}
	state.active.Store(true)
	scope := GlobalScope{state: state}
	authority, err := MintGlobalOwnerAuthority(scope, "billing")
	if err != nil {
		t.Fatal(err)
	}
	if authority.ScopeKind() != "global" || authority.ScopeID() != environmentID ||
		authority.GroupID() != "billing" || authority.GroupGeneration() != 5 ||
		authority.Domain() != "billing" || authority.Generation() != 7 ||
		!authority.OwnsGlobalScope(scope) || authority.OwnsScope(TenantScope{}) {
		t.Fatalf("unexpected global owner authority: %#v", authority)
	}
	state.active.Store(false)
	if authority.OwnsGlobalScope(scope) || authority.ScopeID() != uuid.Nil {
		t.Fatal("global authority survived scope closure")
	}
}

func TestMintGlobalOwnerAuthorityRejectsPublicWrongAndClosedScopes(t *testing.T) {
	if _, err := MintGlobalOwnerAuthority(GlobalScope{}, "billing"); !errors.Is(
		err, ErrScopeClosed,
	) {
		t.Fatalf("zero global scope error = %v", err)
	}
	state := &globalScopeState{environmentID: uuid.New(), mode: SnapshotQuery}
	state.active.Store(true)
	if _, err := MintGlobalOwnerAuthority(
		GlobalScope{state: state}, "billing",
	); !errors.Is(err, ErrScopeOwner) {
		t.Fatalf("public snapshot mint error = %v", err)
	}
	state.mode = Command
	if _, err := MintGlobalOwnerAuthority(
		GlobalScope{state: state}, "forms",
	); !errors.Is(err, ErrScopeOwner) {
		t.Fatalf("wrong global owner error = %v", err)
	}
}
