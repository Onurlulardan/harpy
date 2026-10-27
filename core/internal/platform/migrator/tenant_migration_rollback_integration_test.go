//go:build integration

package migrator

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func TestTenantMigrationPostgreSQL184RollsBackDDLAndRetriesNewRun(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	fixture := newTenantMigrationFixture(t, ctx)
	runner := fixture.runner(t, postgresv1.DatabaseOriginFresh)
	target := postgresstore.TenantMigrationTarget{
		TenantID: "14141414-0000-4000-8000-000000000010", SchemaName: "tenant_rollback",
	}
	insertTenantTarget(t, ctx, fixture.database, target)
	const filename = "20260727110000_retry_probe.fresh.sql"
	content := []byte(`-- +goose Up
CREATE TABLE retry_ddl_probe(id bigint PRIMARY KEY);
INSERT INTO retry_ddl_probe(id) SELECT id FROM retry_gate;
-- +goose Down
DROP TABLE retry_ddl_probe;
`)
	digest := sha256.Sum256(content)
	files := runner.tenantFiles.(fstest.MapFS)
	files[filename] = &fstest.MapFile{Data: content}
	selected := postgresv1.SelectedMigration{
		Owner: "tenancy", Scope: "tenant", Version: 20260727110000,
		TransactionMode: "transactional", Phase: "expand",
		TargetDatabaseGeneration: 2, NextMinimumRuntimeGeneration: 1,
		Origin: postgresv1.DatabaseOriginFresh,
		Path:   "migrations/tenant/" + filename, SHA256: hex.EncodeToString(digest[:]),
	}
	runner.tenantSelected = append(runner.tenantSelected, selected)
	if _, err := runner.executeTenantTarget(ctx, target, tenantRunOne); err == nil ||
		!strings.Contains(err.Error(), "retry_gate") {
		t.Fatalf("expected transactional DDL failure, got %v", err)
	}
	assertTenantFailedDDL(t, ctx, fixture, target, selected.Version)
	qualified := quoteIntegrationIdentifier(target.SchemaName)
	if _, err := fixture.database.ExecContext(ctx,
		"CREATE TABLE "+qualified+".retry_gate(id bigint PRIMARY KEY)",
	); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.database.ExecContext(ctx,
		"INSERT INTO "+qualified+".retry_gate(id) VALUES(1)",
	); err != nil {
		t.Fatal(err)
	}
	if _, err := runner.executeTenantTarget(ctx, target, tenantRunTwo); err != nil {
		t.Fatal(err)
	}
	assertTenantRetrySucceeded(t, ctx, fixture, target, selected.Version)
}

func assertTenantFailedDDL(
	t *testing.T, ctx context.Context, fixture tenantMigrationFixture,
	target postgresstore.TenantMigrationTarget, version int64,
) {
	t.Helper()
	var relation *string
	if err := fixture.database.QueryRowContext(ctx,
		"SELECT pg_catalog.to_regclass($1)::text", target.SchemaName+".retry_ddl_probe",
	).Scan(&relation); err != nil || relation != nil {
		t.Fatalf("rolled back relation = %v err=%v", relation, err)
	}
	var state, code string
	if err := fixture.database.QueryRowContext(ctx, `SELECT state,error_code
FROM platform.tenant_migration_checkpoint_attempts
WHERE tenant_id=$1 AND logical_version=$2 AND attempt=1`,
		target.TenantID, version).Scan(&state, &code); err != nil ||
		state != "failed" || code != "migration_execution_failed" {
		t.Fatalf("failed receipt = %s %s err=%v", state, code, err)
	}
}

func assertTenantRetrySucceeded(
	t *testing.T, ctx context.Context, fixture tenantMigrationFixture,
	target postgresstore.TenantMigrationTarget, version int64,
) {
	t.Helper()
	var attempt int
	var state string
	if err := fixture.database.QueryRowContext(ctx, `SELECT attempt,state
FROM platform.tenant_migration_checkpoints WHERE tenant_id=$1 AND logical_version=$2`,
		target.TenantID, version).Scan(&attempt, &state); err != nil ||
		attempt != 2 || state != "succeeded" {
		t.Fatalf("retry checkpoint = attempt %d state %s err=%v", attempt, state, err)
	}
	var receipts int
	if err := fixture.database.QueryRowContext(ctx, `SELECT count(*)
FROM platform.tenant_migration_checkpoint_attempts
WHERE tenant_id=$1 AND logical_version=$2`, target.TenantID, version).Scan(&receipts); err != nil || receipts != 2 {
		t.Fatalf("retry receipt count = %d err=%v", receipts, err)
	}
}
