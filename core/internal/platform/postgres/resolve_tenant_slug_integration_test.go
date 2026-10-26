//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func assertTenantSlugLocator(
	t *testing.T,
	ctx context.Context,
	pool *Pool,
	database runtimeTestDatabase,
) {
	t.Helper()
	active := uuid.MustParse("71000000-0000-4000-8000-000000000001")
	inactive := uuid.MustParse("71000000-0000-4000-8000-000000000002")
	deleted := uuid.MustParse("71000000-0000-4000-8000-000000000003")
	_, err := database.admin.Exec(ctx, `
INSERT INTO platform.tenants(id,slug,schema_name,status,deleted_at) VALUES
($1,'acme','t_acme','ACTIVE',NULL),
($2,'paused','t_paused','SUSPENDED',NULL),
($3,'deleted','t_deleted','ACTIVE',pg_catalog.now())`, active, inactive, deleted)
	if err != nil {
		t.Fatalf("seed tenant slug locator: %v", err)
	}
	resolved, err := pool.ResolveActiveTenantIDBySlug(ctx, "acme")
	if err != nil || resolved != active {
		t.Fatalf("resolved tenant = %s, %v", resolved, err)
	}
	for _, slug := range []string{"paused", "deleted"} {
		if _, err := pool.ResolveActiveTenantIDBySlug(ctx, slug); !errors.Is(err, ErrTenantInactive) {
			t.Fatalf("inactive tenant %q error = %v", slug, err)
		}
	}
	if _, err := pool.ResolveActiveTenantIDBySlug(ctx, "missing"); !errors.Is(err, ErrTenantMissing) {
		t.Fatalf("missing tenant error = %v", err)
	}
}
