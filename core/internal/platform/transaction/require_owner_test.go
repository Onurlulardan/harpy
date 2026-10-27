package transaction

import (
	"errors"
	"testing"
)

func TestRequireOwnerAcceptsOnlyExactOperationOwner(t *testing.T) {
	operation, err := NewSnapshotAuthorityIssuer().Issue(
		"restquery.execute", map[string]int64{"forms": 7},
	)
	if err != nil {
		t.Fatal(err)
	}
	state := &scopeState{operation: operation, mode: SnapshotQuery}
	state.active.Store(true)
	scope := TenantScope{state: state}
	if err := RequireOwner(scope, "forms"); err != nil {
		t.Fatalf("exact owner rejected: %v", err)
	}
	for _, owner := range []string{"crm", "Forms", "", "forms "} {
		if err := RequireOwner(scope, owner); !errors.Is(err, ErrScopeOwner) {
			t.Fatalf("owner %q error=%v", owner, err)
		}
	}
}

func TestRequireOwnerRejectsClosedOrTamperedScope(t *testing.T) {
	operation, err := NewSnapshotAuthorityIssuer().Issue(
		"restquery.execute", map[string]int64{"forms": 1},
	)
	if err != nil {
		t.Fatal(err)
	}
	state := &scopeState{operation: operation, mode: SnapshotQuery}
	scope := TenantScope{state: state}
	if err := RequireOwner(scope, "forms"); !errors.Is(err, ErrScopeClosed) {
		t.Fatalf("closed scope error=%v", err)
	}
	state.active.Store(true)
	state.mode = Command
	if err := RequireOwner(scope, "forms"); !errors.Is(err, ErrScopeOwner) {
		t.Fatalf("mode-tampered scope error=%v", err)
	}
	state.mode = SnapshotQuery
	state.operation.issued = false
	if err := RequireOwner(scope, "forms"); !errors.Is(err, ErrScopeOwner) {
		t.Fatalf("operation-tampered scope error=%v", err)
	}
}
