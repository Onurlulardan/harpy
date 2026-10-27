//go:build integration

package transaction

import (
	"context"
	"errors"
	"testing"
)

func assertGlobalScopeNesting(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
) {
	t.Helper()
	globalRequest := globalPublicRequestFor(t, 5, 7)
	tenantRequest := scopeSnapshotRequestFor(t, tenantA, map[string]int64{"forms": 1})
	err := coordinator.InGlobal(
		ctx, globalRequest, func(ctx context.Context, _ GlobalScope) error {
			if err := coordinator.InGlobal(
				ctx, globalRequest, func(context.Context, GlobalScope) error { return nil },
			); !errors.Is(err, ErrNestedScope) {
				t.Fatalf("nested global scope error = %v", err)
			}
			return coordinator.InTenant(
				ctx, tenantRequest, func(context.Context, TenantScope) error { return nil },
			)
		},
	)
	if !errors.Is(err, ErrNestedScope) {
		t.Fatalf("global to tenant nesting error = %v", err)
	}
	err = coordinator.InTenant(
		ctx, tenantRequest, func(ctx context.Context, _ TenantScope) error {
			return coordinator.InGlobal(
				ctx, globalRequest, func(context.Context, GlobalScope) error { return nil },
			)
		},
	)
	if !errors.Is(err, ErrNestedScope) {
		t.Fatalf("tenant to global nesting error = %v", err)
	}
}
