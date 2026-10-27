//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

const (
	tenantExecutorReleaseID = "20000000-0000-4000-8000-000000000099"
	tenantExecutorDeployRun = "30000000-0000-4000-8000-000000000099"
)

func prepareTenantExecutorFoundation(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	digest := func(character string) string { return strings.Repeat(character, 64) }
	statements := []string{
		`CREATE TABLE platform.migration_runs (
run_id uuid PRIMARY KEY, release_id uuid NOT NULL, environment_id uuid NOT NULL,
database_origin text NOT NULL, command text NOT NULL, state text NOT NULL,
migration_manifest_sha256 text NOT NULL, extension_manifest_sha256 text NOT NULL,
roles_manifest_sha256 text NOT NULL, privileges_manifest_sha256 text NOT NULL,
ownership_manifest_sha256 text NOT NULL, created_at timestamptz NOT NULL DEFAULT now(),
started_at timestamptz, finished_at timestamptz, failure_code text,
CONSTRAINT migration_runs_command_check CHECK (command IN
('plan','verify','up','schema_executor')))`,
		`CREATE TABLE platform.release_generations (
migration_run_id uuid PRIMARY KEY, release_id uuid NOT NULL, environment_id uuid NOT NULL,
database_origin text NOT NULL, database_generation bigint NOT NULL,
runtime_generation bigint NOT NULL, migration_manifest_sha256 text NOT NULL)`,
		`CREATE TABLE platform.schema_contracts (
component text PRIMARY KEY, status text NOT NULL, verified_run_id uuid NOT NULL,
database_generation bigint NOT NULL, migration_manifest_sha256 text NOT NULL)`,
		`CREATE TABLE platform.tenant_migration_requirement_catalogs (
migration_run_id uuid PRIMARY KEY, contract_version text NOT NULL,
release_id uuid NOT NULL, environment_id uuid NOT NULL, database_origin text NOT NULL,
migration_manifest_sha256 text NOT NULL, catalog_sha256 text NOT NULL,
requirement_count integer NOT NULL)`,
		`CREATE TABLE platform.tenant_migration_requirements (
migration_run_id uuid NOT NULL, ordinal integer NOT NULL, owner text NOT NULL,
logical_version bigint NOT NULL, database_origin text NOT NULL,
selected_path text NOT NULL, selected_sha256 text NOT NULL)`,
		`CREATE TABLE platform.tenant_migration_checkpoints (
tenant_id uuid NOT NULL, owner text NOT NULL, logical_version bigint NOT NULL,
database_origin text NOT NULL, selected_path text NOT NULL, selected_sha256 text NOT NULL,
run_id uuid NOT NULL, state text NOT NULL, attempt integer NOT NULL,
started_at timestamptz, finished_at timestamptz, error_code text, updated_at timestamptz)`,
	}
	for _, statement := range statements {
		if _, err := database.ExecContext(ctx, statement); err != nil {
			t.Fatalf("create executor foundation: %v", err)
		}
	}
	applyPlatformExecutorMigration(t, ctx, database,
		"20260722030000_allow_tenant_migration_executor_runs."+origin+".sql", origin)
	applyPlatformExecutorMigration(t, ctx, database,
		"20260722030001_create_new_tenant_checkpoint_verifier."+origin+".sql", origin)
	execTenantExecutorFixture(t, ctx, database, `INSERT INTO platform.migration_runs (
run_id,release_id,environment_id,database_origin,command,state,
migration_manifest_sha256,extension_manifest_sha256,roles_manifest_sha256,
privileges_manifest_sha256,ownership_manifest_sha256,started_at,finished_at)
VALUES ($1,$2,$3,$4,'up','succeeded',$5,$6,$7,$8,$9,now(),now())`,
		tenantExecutorDeployRun, tenantExecutorReleaseID, testEnvironmentID, origin,
		digest("a"), digest("b"), digest("c"), digest("d"), digest("e"))
	execTenantExecutorFixture(t, ctx, database,
		`INSERT INTO platform.release_generations VALUES ($1,$2,$3,$4,7,11,$5)`,
		tenantExecutorDeployRun, tenantExecutorReleaseID, testEnvironmentID,
		origin, digest("a"))
	execTenantExecutorFixture(t, ctx, database,
		`INSERT INTO platform.schema_contracts VALUES ('core','ready',$1,7,$2)`,
		tenantExecutorDeployRun, digest("a"))
	execTenantExecutorFixture(t, ctx, database, `INSERT INTO
platform.tenant_migration_requirement_catalogs VALUES
($1,'tenant-migration-requirement-catalog.v2',$2,$3,$4,$5,$6,1)`,
		tenantExecutorDeployRun, tenantExecutorReleaseID, testEnvironmentID,
		origin, digest("a"), digest("b"))
	execTenantExecutorFixture(t, ctx, database, `INSERT INTO
platform.tenant_migration_requirements VALUES
($1,1,'tenancy',20260714011000,$2,'tenancy/new.sql',$3)`,
		tenantExecutorDeployRun, origin, digest("c"))
}

func execTenantExecutorFixture(
	t *testing.T, ctx context.Context, database *sql.DB,
	statement string, arguments ...any,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, statement, arguments...); err != nil {
		t.Fatalf("seed executor foundation: %v", err)
	}
}

func applyPlatformExecutorMigration(
	t *testing.T, ctx context.Context, database *sql.DB, name, origin string,
) {
	t.Helper()
	path := filepath.Join("..", "..", "..", "..", "platform", "migrator",
		"migrations", origin, name)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, string(content)); err != nil {
		t.Fatalf("apply %s: %v", name, err)
	}
}

func acceptTenantExecutorOperation(
	t *testing.T, ctx context.Context, database *sql.DB,
) onboardingFixture {
	t.Helper()
	return acceptTenantExecutorOperationMode(t, ctx, database, "onpremise")
}

func acceptTenantExecutorOperationMode(
	t *testing.T, ctx context.Context, database *sql.DB, mode string,
) onboardingFixture {
	t.Helper()
	fixture := newOnboardingFixture(t, mode)
	transaction := beginRuntimeTransaction(t, ctx, database)
	arguments := fixture.acceptArguments(fixture.manifestSHA)
	arguments[3] = fixture.operationID.String()
	placeholders := "$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16," +
		"$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31"
	var operationID uuid.UUID
	var state string
	var revision int64
	var updatedAt time.Time
	err := transaction.QueryRowContext(ctx,
		"SELECT * FROM platform.tenancy_onboarding_accept_v2("+placeholders+")",
		arguments...).Scan(&operationID, &state, &revision, &updatedAt)
	if err != nil || operationID != fixture.operationID || state != "accepted" {
		_ = transaction.Rollback()
		t.Fatalf("accept tenant executor operation: %s/%s %v", operationID, state, err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return fixture
}

type executorClaim struct {
	jobID, tenantID, runID, previousRunID, leaseID, profileID uuid.UUID
	operationID                                               uuid.UUID
	schema                                                    string
	attempt                                                   int
}
