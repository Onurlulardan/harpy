//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy/domain"
)

type settingsRoutineResult struct {
	revision int64
	replay   bool
	outcome  string
	err      error
}

func TestTenantSettingsRuntimeAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()
	database := startTenantMigrationDatabase(t, ctx)
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			resetTenantMigrationCatalog(t, ctx, database)
			primary := prepareSettingsOrigin(t, ctx, database, origin)
			secondary := "22222222-2222-4222-8222-222222222222"
			insertSettingsTenant(t, ctx, database, secondary, "Secondary")
			assertSettingsPrivileges(t, ctx, database)
			assertSettingsReplayAndIsolation(t, ctx, database, primary, secondary)
			assertSettingsConcurrency(t, ctx, database, primary)
			assertSettingsV1Fence(t, ctx, database, primary)
			assertSettingsReceiptImmutable(t, ctx, database, primary)
		})
	}
}

func prepareSettingsOrigin(t *testing.T, ctx context.Context,
	database *sql.DB, origin string,
) string {
	t.Helper()
	if origin == "fresh" {
		runPlatformMigrations(t, ctx, database, origin)
		insertFreshTenant(t, ctx, database)
		return freshTenantID
	}
	prepareV1TenantRegistry(t, ctx, database)
	runPlatformMigrations(t, ctx, database, origin)
	return adoptedTenantID
}

func insertSettingsTenant(t *testing.T, ctx context.Context,
	database *sql.DB, tenantID, name string,
) {
	t.Helper()
	schema, err := domain.SchemaNameFromUUID(tenantID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.ExecContext(ctx, `INSERT INTO platform.tenants(
id,name,slug,status,settings,owner_id,schema_name,created_at,updated_at
) VALUES($1,$2,$3,'ACTIVE','{}',$4,$5,$6,$6)`, tenantID, name,
		"tenant-"+tenantID[:8], testOwnerID, schema.String(), time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
}

func assertSettingsPrivileges(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	var tableRead, receiptWrite, readRoutine, updateRoutine bool
	err := database.QueryRowContext(ctx, `SELECT
has_table_privilege('harpy_runtime','platform.tenants','SELECT'),
has_table_privilege('harpy_runtime','platform.tenancy_settings_operation_receipts','INSERT'),
has_function_privilege('harpy_runtime','platform.tenancy_settings_read_v1(uuid)','EXECUTE'),
has_function_privilege('harpy_runtime','platform.tenancy_settings_update_v1(uuid,text,bigint,text,bytea)','EXECUTE')
`).Scan(&tableRead, &receiptWrite, &readRoutine, &updateRoutine)
	if err != nil || tableRead || receiptWrite || !readRoutine || !updateRoutine {
		t.Fatalf("settings privileges table=%t receipt=%t read=%t update=%t err=%v",
			tableRead, receiptWrite, readRoutine, updateRoutine, err)
	}
}
