//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jackc/pgx/v5"
)

func stageTenantFinalizationRace(
	t *testing.T,
	ctx context.Context,
) (*sql.DB, *Runner, *sql.Conn) {
	t.Helper()
	fixture := newTenantMigrationFixture(t, ctx)
	runner := newRunnerIntegrationFixture(t, ctx, fixture.database, fixture.config)
	tenantRunner := fixture.runner(t, runner.identity.databaseOrigin)
	runner.tenantFiles = tenantRunner.tenantFiles
	runner.tenantSelected = tenantRunner.tenantSelected
	runner.tenantDatabase = tenantRunner.tenantDatabase
	connection, err := fixture.database.Conn(ctx)
	if err != nil {
		t.Fatalf("open race finalization connection: %v", err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	if _, err := runner.ensureRunRunning(ctx, connection, runnerIntegrationRunID); err != nil {
		t.Fatalf("start race migration run: %v", err)
	}
	return fixture.database, runner, connection
}

func insertRaceTenant(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
	tenantID, schemaName string,
) {
	t.Helper()
	if _, err := transaction.ExecContext(
		ctx, "CREATE SCHEMA "+pgx.Identifier{schemaName}.Sanitize()+
			" AUTHORIZATION harpy_owner",
	); err != nil {
		t.Fatalf("create race tenant schema: %v", err)
	}
	_, err := transaction.ExecContext(ctx, `INSERT INTO platform.tenants
(id,name,slug,status,settings,owner_id,schema_name,created_at,updated_at)
VALUES($1,$2::varchar(255),$2::varchar(100),'ACTIVE','{}',$3,$2::text,
pg_catalog.clock_timestamp(),
pg_catalog.clock_timestamp())`, tenantID, schemaName, tenantOwner)
	if err != nil {
		t.Fatalf("insert race tenant: %v", err)
	}
}
