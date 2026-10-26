package transaction

import (
	"context"
	"testing"
)

func TestScopeContextMarksNetworkAsForbidden(t *testing.T) {
	if !NetworkAllowed(context.Background()) {
		t.Fatal("ordinary context unexpectedly forbids network")
	}
	scope := TenantScope{state: &scopeState{}}
	if NetworkAllowed(markScopeContext(context.Background(), scope)) {
		t.Fatal("tenant transaction context allowed network")
	}
}

func TestScopeContextOwnsOnlyExactScopeState(t *testing.T) {
	first := TenantScope{state: &scopeState{}}
	copyOfFirst := first
	second := TenantScope{state: &scopeState{}}
	ctx := markScopeContext(context.Background(), first)
	derived, cancel := context.WithCancel(ctx)
	defer cancel()
	if !ContextOwnsScope(ctx, first) || !ContextOwnsScope(ctx, copyOfFirst) {
		t.Fatal("exact scope state was not recognized")
	}
	if !ContextOwnsScope(derived, first) {
		t.Fatal("derived context lost exact scope ownership")
	}
	if ContextOwnsScope(ctx, second) || ContextOwnsScope(ctx, TenantScope{}) ||
		ContextOwnsScope(context.Background(), first) || ContextOwnsScope(nil, first) {
		t.Fatal("foreign or missing scope context was accepted")
	}
}
