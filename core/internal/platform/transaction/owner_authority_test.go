package transaction

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestMintOwnerAuthorityProjectsExactVerifiedDomain(t *testing.T) {
	tenantID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	operation, err := NewAuthorityIssuer().Issue(
		"forms.records.create", "billing_forms", 7,
		map[string]int64{"billing": 3, "forms": 5},
	)
	if err != nil {
		t.Fatal(err)
	}
	scope := activeAuthorityScope(tenantID, operation)
	authority, err := MintOwnerAuthority(scope, "forms")
	if err != nil {
		t.Fatal(err)
	}
	if authority.ScopeKind() != "tenant" || authority.ScopeID() != tenantID ||
		authority.GroupID() != "billing_forms" || authority.GroupGeneration() != 7 ||
		authority.Domain() != "forms" || authority.Generation() != 5 ||
		authority.Origin() != WriterOriginV2 || !authority.OwnsScope(scope) {
		t.Fatalf("unexpected owner authority: %#v", authority)
	}
	other := activeAuthorityScope(tenantID, operation)
	if authority.OwnsScope(other) {
		t.Fatal("authority crossed an exact scope identity")
	}
	scope.state.active.Store(false)
	if authority.OwnsScope(scope) {
		t.Fatal("authority survived scope closure")
	}
	if authority.ScopeKind() != "" || authority.ScopeID() != uuid.Nil ||
		authority.GroupID() != "" || authority.GroupGeneration() != 0 ||
		authority.Domain() != "" || authority.Generation() != 0 ||
		authority.Origin() != "" {
		t.Fatal("closed authority leaked its projection")
	}
}

func TestMintOwnerAuthorityRejectsUnownedSnapshotAndClosedScopes(t *testing.T) {
	tenantID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	if _, err := MintOwnerAuthority(TenantScope{}, "iam"); !errors.Is(err, ErrScopeClosed) {
		t.Fatalf("zero scope error = %v", err)
	}
	command, _ := NewAuthorityIssuer().Issue(
		"iam.users.create", "iam_system_resources", 1,
		map[string]int64{"iam": 1, "system_resource_layout": 1},
	)
	commandScope := activeAuthorityScope(tenantID, command)
	if _, err := MintOwnerAuthority(commandScope, "forms"); !errors.Is(err, ErrScopeOwner) {
		t.Fatalf("wrong owner error = %v", err)
	}
	commandScope.state.active.Store(false)
	if _, err := MintOwnerAuthority(commandScope, "iam"); !errors.Is(err, ErrScopeClosed) {
		t.Fatalf("closed scope error = %v", err)
	}
	snapshot, _ := NewSnapshotAuthorityIssuer().Issue(
		"restquery.execute", map[string]int64{"iam": 1},
	)
	if _, err := MintOwnerAuthority(activeAuthorityScope(tenantID, snapshot), "iam"); !errors.Is(err, ErrScopeOwner) {
		t.Fatalf("snapshot scope error = %v", err)
	}
}

func activeAuthorityScope(tenantID uuid.UUID, operation VerifiedOperation) TenantScope {
	state := &scopeState{tenantID: tenantID, operation: operation, mode: operation.mode}
	state.active.Store(true)
	return TenantScope{state: state}
}
