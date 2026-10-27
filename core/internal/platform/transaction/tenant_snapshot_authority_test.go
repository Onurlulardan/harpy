package transaction

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestTenantSnapshotAuthorityProjectsExactOwnerFence(t *testing.T) {
	tenantID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	operation, err := NewSnapshotAuthorityIssuer().Issue(
		"analytics.query.execute", map[string]int64{"analytics": 3, "forms": 7},
	)
	if err != nil {
		t.Fatal(err)
	}
	scope := activeAuthorityScope(tenantID, operation)
	authority, err := MintTenantSnapshotAuthority(scope, "forms")
	if err != nil {
		t.Fatal(err)
	}
	if authority.TenantID() != tenantID || authority.Domain() != "forms" ||
		authority.Generation() != 7 || !authority.Owns(scope) {
		t.Fatalf("unexpected snapshot authority: %#v", authority)
	}
	other := activeAuthorityScope(tenantID, operation)
	if authority.Owns(other) {
		t.Fatal("snapshot authority crossed scope identity")
	}
	scope.state.active.Store(false)
	if authority.Owns(scope) || authority.TenantID() != uuid.Nil ||
		authority.Domain() != "" || authority.Generation() != 0 {
		t.Fatal("closed snapshot authority leaked its fence")
	}
}

func TestTenantSnapshotAuthorityRejectsWrongModeOwnerAndClosedScope(t *testing.T) {
	tenantID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	if _, err := MintTenantSnapshotAuthority(TenantScope{}, "forms"); !errors.Is(err, ErrScopeClosed) {
		t.Fatalf("zero scope error = %v", err)
	}
	command, _ := NewAuthorityIssuer().Issue(
		"billing.subscription.create", "billing", 1,
		map[string]int64{"billing": 1},
	)
	if _, err := MintTenantSnapshotAuthority(
		activeAuthorityScope(tenantID, command), "forms",
	); !errors.Is(err, ErrScopeOwner) {
		t.Fatalf("command scope error = %v", err)
	}
	snapshot, _ := NewSnapshotAuthorityIssuer().Issue(
		"analytics.query.execute", map[string]int64{"analytics": 1},
	)
	if _, err := MintTenantSnapshotAuthority(
		activeAuthorityScope(tenantID, snapshot), "forms",
	); !errors.Is(err, ErrScopeOwner) {
		t.Fatalf("wrong owner error = %v", err)
	}
}
