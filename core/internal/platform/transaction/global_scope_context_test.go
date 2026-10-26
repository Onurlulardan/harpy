package transaction

import (
	"context"
	"testing"
)

func TestGlobalScopeContextOwnsOnlyExactStateAndForbidsNesting(t *testing.T) {
	first := GlobalScope{state: &globalScopeState{}}
	copyOfFirst := first
	second := GlobalScope{state: &globalScopeState{}}
	ctx := markGlobalScopeContext(context.Background(), first)
	if !ContextOwnsGlobalScope(ctx, first) ||
		!ContextOwnsGlobalScope(ctx, copyOfFirst) {
		t.Fatal("exact global scope state was not recognized")
	}
	if ContextOwnsGlobalScope(ctx, second) ||
		ContextOwnsGlobalScope(ctx, GlobalScope{}) ||
		ContextOwnsGlobalScope(context.Background(), first) {
		t.Fatal("foreign global scope state was accepted")
	}
	if NetworkAllowed(ctx) || !nestedScope(ctx) {
		t.Fatal("global transaction context allowed network or nesting")
	}
}
