//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const adoptedTenantID = "11111111-1111-4111-8111-111111111111"

func prepareV1TenantRegistry(t *testing.T, ctx context.Context, database *sql.DB) int64 {
	t.Helper()
	_, err := database.ExecContext(ctx, `
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
)`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx,
		"CREATE UNIQUE INDEX idx_tenants_slug ON public.tenants (slug)"); err != nil {
		t.Fatal(err)
	}
	_, err = database.ExecContext(ctx, `
INSERT INTO public.tenants (
    id, name, slug, status, settings, owner_id, created_at, updated_at
) VALUES ($1, 'Adopted', 'acme', 'ACTIVE', NULL, $2,
    TIMESTAMP '2026-07-13 09:30:00', TIMESTAMP '2026-07-13 09:30:00')`,
		adoptedTenantID, testOwnerID)
	if err != nil {
		t.Fatal(err)
	}
	var oid int64
	if err := database.QueryRowContext(ctx,
		"SELECT 'public.tenants'::regclass::oid::bigint").Scan(&oid); err != nil {
		t.Fatal(err)
	}
	return oid
}

func assertAdoptedTenant(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	originalOID int64,
) {
	t.Helper()
	var movedOID int64
	var schemaName string
	var settingsNormalized, timestampConverted bool
	err := database.QueryRowContext(ctx, `
SELECT 'platform.tenants'::regclass::oid::bigint,
       schema_name,
       settings = '{}'::jsonb,
       created_at = TIMESTAMPTZ '2026-07-13 09:30:00+00'
FROM platform.tenants WHERE id = $1`, adoptedTenantID).
		Scan(&movedOID, &schemaName, &settingsNormalized, &timestampConverted)
	if err != nil {
		t.Fatal(err)
	}
	if movedOID != originalOID || schemaName != "tenant_acme" ||
		!settingsNormalized || !timestampConverted {
		t.Fatalf("adoption mismatch oid=%d/%d schema=%q settings=%v timestamp=%v",
			movedOID, originalOID, schemaName, settingsNormalized, timestampConverted)
	}
}
