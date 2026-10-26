//go:build integration

package transaction

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

const missingTenant = "cccccccc-cccc-4ccc-8ccc-cccccccccccc"

func assertScopeDenials(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	binder postgres.OwnerBinder[scopeProbe],
	database scopeTestDatabase,
) {
	t.Helper()
	assertScopeError(t, ctx, coordinator, scopeRequestFor(t, missingTenant, 1), ErrTenantMissing)
	assertInactiveTenant(t, ctx, coordinator, database)
	assertMismatchedIdentity(t, ctx, coordinator, database)
	assertScopeError(t, ctx, coordinator, scopeRequestFor(t, tenantA, 2), ErrWriterAuthority)
	assertMissingGroupMember(t, ctx, coordinator, database)
	assertRuntimeDDLDenied(t, ctx, coordinator, binder, database)
}

func assertScopeError(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	request ScopeRequest,
	want error,
) {
	t.Helper()
	err := coordinator.InTenant(ctx, request, func(context.Context, TenantScope) error { return nil })
	if !errors.Is(err, want) {
		t.Fatalf("scope error = %v, want %v", err, want)
	}
}

func assertInactiveTenant(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	database scopeTestDatabase,
) {
	t.Helper()
	adminExecute(t, ctx, database, "UPDATE platform.tenants SET status='SUSPENDED' WHERE id=$1", tenantA)
	t.Cleanup(func() {
		adminExecute(t, context.Background(), database, "UPDATE platform.tenants SET status='ACTIVE' WHERE id=$1", tenantA)
	})
	assertScopeError(t, ctx, coordinator, scopeRequestFor(t, tenantA, 1), ErrTenantInactive)
	adminExecute(t, ctx, database, "UPDATE platform.tenants SET status='ACTIVE' WHERE id=$1", tenantA)
}

func assertMismatchedIdentity(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	database scopeTestDatabase,
) {
	t.Helper()
	adminExecute(t, ctx, database, "UPDATE tenant_a.tenancy_schema_identity SET tenant_id=$1", tenantB)
	t.Cleanup(func() {
		adminExecute(t, context.Background(), database, "UPDATE tenant_a.tenancy_schema_identity SET tenant_id=$1", tenantA)
	})
	assertScopeError(t, ctx, coordinator, scopeRequestFor(t, tenantA, 1), ErrTenantIdentity)
	adminExecute(t, ctx, database, "UPDATE tenant_a.tenancy_schema_identity SET tenant_id=$1", tenantA)
}

func adminExecute(
	t *testing.T,
	ctx context.Context,
	database scopeTestDatabase,
	statement string,
	arguments ...any,
) {
	t.Helper()
	if _, err := database.admin.Exec(ctx, statement, arguments...); err != nil {
		t.Fatalf("admin statement failed: %v", err)
	}
}
