//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"testing"
)

func prepareRunnerV1TenantRegistry(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `
CREATE SCHEMA tenant_adopted AUTHORIZATION harpy_owner;
CREATE TABLE public.tenants (
    id uuid PRIMARY KEY DEFAULT pg_catalog.gen_random_uuid(),
    name varchar(255) NOT NULL,
    slug varchar(100) NOT NULL,
    status varchar(20) NOT NULL DEFAULT 'ACTIVE',
    settings jsonb DEFAULT '{}',
    owner_id uuid NOT NULL,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    deleted_at timestamp without time zone
);
CREATE UNIQUE INDEX idx_tenants_slug ON public.tenants (slug);
INSERT INTO public.tenants (
    id, name, slug, status, settings, owner_id, created_at, updated_at
) VALUES (
    '018f0000-0000-7000-8000-000000000781',
    'Adopted', 'adopted', 'ACTIVE', NULL,
    '018f0000-0000-7000-8000-000000000782',
    TIMESTAMP '2026-07-13 09:30:00',
    TIMESTAMP '2026-07-13 09:30:00'
)`)
	if err != nil {
		t.Fatalf("prepare V1 tenant registry: %v", err)
	}
}
