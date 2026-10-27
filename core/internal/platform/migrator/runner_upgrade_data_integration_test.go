//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

const rollingUpgradeTenantID = "018f0000-0000-7000-8000-000000000211"
const rollingUpgradeOwnerID = "018f0000-0000-7000-8000-000000000212"

func seedRollingUpgradeTenant(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	if _, err := database.ExecContext(
		ctx, "CREATE SCHEMA tenant_preserved AUTHORIZATION harpy_owner",
	); err != nil {
		t.Fatalf("create rolling upgrade tenant schema: %v", err)
	}
	_, err := database.ExecContext(ctx, `INSERT INTO platform.tenants
(id,name,slug,status,settings,owner_id,schema_name,created_at,updated_at)
VALUES($1,'Preserved tenant','preserved-tenant','ACTIVE',
'{"release":"generation-one"}',$2,'tenant_preserved',
TIMESTAMPTZ '2026-07-14 10:00:00+00',TIMESTAMPTZ '2026-07-14 10:00:00+00')`,
		rollingUpgradeTenantID, rollingUpgradeOwnerID)
	if err != nil {
		t.Fatalf("seed rolling upgrade tenant: %v", err)
	}
}

func assertRollingUpgradeTenantPreserved(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	var name, slug, status, settings, ownerID, schemaName string
	var createdAt time.Time
	err := database.QueryRowContext(ctx, `SELECT name,slug,status,
settings::text,owner_id::text,schema_name,created_at
FROM platform.tenants WHERE id=$1`, rollingUpgradeTenantID).Scan(
		&name, &slug, &status, &settings, &ownerID, &schemaName, &createdAt,
	)
	if err != nil {
		t.Fatalf("read preserved rolling upgrade tenant: %v", err)
	}
	if name != "Preserved tenant" || slug != "preserved-tenant" ||
		status != "ACTIVE" || settings != `{"release": "generation-one"}` ||
		ownerID != rollingUpgradeOwnerID || schemaName != "tenant_preserved" ||
		!createdAt.Equal(time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("rolling upgrade changed tenant data: %s %s %s %s %s %s %s",
			name, slug, status, settings, ownerID, schemaName, createdAt)
	}
}
