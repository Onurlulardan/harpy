package transaction

import (
	"errors"
	"testing"
)

func TestRequireActiveAcceptsOnlyMatchingLiveScope(t *testing.T) {
	state := &scopeState{mode: SnapshotQuery}
	state.active.Store(true)
	scope := TenantScope{state: state}

	if err := RequireActive(scope, SnapshotQuery); err != nil {
		t.Fatalf("matching live scope: %v", err)
	}
	if err := RequireActive(scope, Command); !errors.Is(err, ErrScopeMode) {
		t.Fatalf("wrong mode error = %v", err)
	}
	if err := RequireActive(scope, ScopeMode("unknown")); !errors.Is(err, ErrScopeMode) {
		t.Fatalf("invalid expected mode error = %v", err)
	}

	state.active.Store(false)
	if err := RequireActive(scope, SnapshotQuery); !errors.Is(err, ErrScopeClosed) {
		t.Fatalf("closed scope error = %v", err)
	}
}

func TestRequireActiveRejectsZeroScope(t *testing.T) {
	if err := RequireActive(TenantScope{}, SnapshotQuery); !errors.Is(err, ErrScopeClosed) {
		t.Fatalf("zero scope error = %v", err)
	}
}
