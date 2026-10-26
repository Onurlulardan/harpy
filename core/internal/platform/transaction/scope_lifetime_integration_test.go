//go:build integration

package transaction

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func assertScopeLifetime(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	binder postgres.OwnerBinder[scopeProbe],
) {
	t.Helper()
	request := scopeRequestFor(t, tenantA, 1)
	var retainedScope TenantScope
	var retainedProbe scopeProbe
	err := coordinator.InTenant(ctx, request, func(ctx context.Context, scope TenantScope) error {
		retainedScope = scope
		probe, err := Bind(scope, binder)
		if err != nil {
			return err
		}
		retainedProbe = probe
		nestedErr := coordinator.InTenant(ctx, request, func(context.Context, TenantScope) error {
			return nil
		})
		if !errors.Is(nestedErr, ErrNestedScope) {
			t.Fatalf("nested scope error = %v", nestedErr)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("lifetime scope: %v", err)
	}
	if _, err := Bind(retainedScope, binder); !errors.Is(err, ErrScopeClosed) {
		t.Fatalf("binding retained scope error = %v", err)
	}
	var count int
	err = retainedProbe.database.QueryRow(ctx, "SELECT count(*) FROM iam_probe").Scan(&count)
	if !errors.Is(err, postgres.ErrTransactionClosed) {
		t.Fatalf("retained executor error = %v", err)
	}
	if _, err := retainedProbe.database.Execute(ctx, "SELECT 1"); !errors.Is(err, postgres.ErrTransactionClosed) {
		t.Fatalf("retained executor execute error = %v", err)
	}
	if _, err := retainedProbe.database.Query(ctx, "SELECT 1"); !errors.Is(err, postgres.ErrTransactionClosed) {
		t.Fatalf("retained executor query error = %v", err)
	}
}
