//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

const verifyTenantRequirementCatalogSQL = `SELECT verified, catalog_sha256,
requirement_count, target_count FROM platform.platform_migrator_verify_tenant_requirement_catalog_v2(
$1,$2,$3,$4)`

func TestTenantRequirementCatalogPostgreSQL184IsReleaseBoundAndFailClosed(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database, config := startRunnerIntegrationDatabase(t, ctx)
	runner := newRunnerIntegrationFixture(t, ctx, database, config)
	if _, err := runner.Up(ctx, UpRequest{RunID: runnerIntegrationRunID}); err != nil {
		t.Fatalf("finalize requirement catalog release: %v", err)
	}
	catalog, err := postgresstore.GetTenantRequirementCatalog(
		ctx, database, runnerIntegrationRunID,
	)
	if err != nil || len(catalog.Requirements) != len(runner.tenantSelected) ||
		len(catalog.Targets) != catalog.TargetCount {
		t.Fatalf("stored requirement catalog = %#v err=%v", catalog, err)
	}
	var ordered bool
	if err := database.QueryRowContext(ctx, `SELECT
release_row.recorded_at <= catalog.recorded_at
AND catalog.recorded_at <= contract_row.verified_at
FROM platform.tenant_migration_requirement_catalogs AS catalog
JOIN platform.release_generations AS release_row USING (migration_run_id)
JOIN platform.schema_contracts AS contract_row
  ON contract_row.verified_run_id = catalog.migration_run_id
WHERE catalog.migration_run_id=$1`, runnerIntegrationRunID).Scan(&ordered); err != nil || !ordered {
		t.Fatalf("requirement finalization order = %t err=%v", ordered, err)
	}
	var digestExact bool
	if err := database.QueryRowContext(ctx, `SELECT catalog_sha256 =
pg_catalog.encode(pg_catalog.sha256(canonical_document),'hex')
FROM platform.tenant_migration_requirement_catalogs WHERE migration_run_id=$1`,
		runnerIntegrationRunID,
	).Scan(&digestExact); err != nil || !digestExact {
		t.Fatalf("database catalog digest exact = %t err=%v", digestExact, err)
	}
	verified, digest, count := readTenantRequirementVerification(
		t, ctx, database, runner, runnerIntegrationRunID,
	)
	if !verified || digest != catalog.CatalogHash || count != len(runner.tenantSelected) {
		t.Fatalf("requirement verification = %t %s %d", verified, digest, count)
	}
	missing, missingDigest, missingCount := readTenantRequirementVerification(
		t, ctx, database, runner, "99999999-0000-4000-8000-000000000099",
	)
	if missing || missingDigest != "" || missingCount != 0 {
		t.Fatalf("missing requirement authority = %t %q %d",
			missing, missingDigest, missingCount)
	}
	assertTenantRequirementPrivilegeBoundary(t, ctx, database)
	assertTenantRequirementReplayAndTamper(t, ctx, database, runner)
	assertTenantRequirementImmutability(t, ctx, database)
	late := postgresstore.TenantMigrationTarget{
		TenantID:   "81818181-0000-4000-8000-000000000081",
		SchemaName: "tenant_requirement_late",
	}
	insertTenantTarget(t, ctx, database, late)
	verified, _, _ = readTenantRequirementVerification(
		t, ctx, database, runner, runnerIntegrationRunID,
	)
	if verified {
		t.Fatal("late active tenant passed requirement verification")
	}
}

func assertTenantRequirementPrivilegeBoundary(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	var execute, summaryRead, detailRead bool
	err := database.QueryRowContext(ctx, `SELECT
pg_catalog.has_function_privilege('harpy_cutover_finalizer',
'platform.platform_migrator_verify_tenant_requirement_catalog_v2(uuid,uuid,uuid,text)',
'EXECUTE'),
pg_catalog.has_table_privilege('harpy_cutover_finalizer',
'platform.tenant_migration_requirement_catalogs','SELECT'),
pg_catalog.has_table_privilege('harpy_cutover_finalizer',
'platform.tenant_migration_requirements','SELECT')`).Scan(
		&execute, &summaryRead, &detailRead,
	)
	if err != nil || execute || summaryRead || detailRead {
		t.Fatalf("tenant requirement privileges = %t/%t/%t err=%v",
			execute, summaryRead, detailRead, err)
	}
}

func readTenantRequirementVerification(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	runner *Runner,
	runID string,
) (bool, string, int) {
	t.Helper()
	var verified bool
	var digest sql.NullString
	var count int
	var targetCount int
	err := database.QueryRowContext(ctx, verifyTenantRequirementCatalogSQL,
		runID, runner.identity.releaseID, runner.identity.environmentID,
		runner.identity.databaseOrigin,
	).Scan(&verified, &digest, &count, &targetCount)
	if err != nil {
		t.Fatalf("verify tenant requirement catalog: %v", err)
	}
	return verified, digest.String, count
}

func assertTenantRequirementReplayAndTamper(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	runner *Runner,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.persistTenantRequirementCatalog(
		ctx, transaction, runnerIntegrationRunID,
	); err != nil {
		t.Fatalf("exact requirement catalog replay: %v", err)
	}
	if err := transaction.Rollback(); err != nil {
		t.Fatal(err)
	}
	tampered := *runner
	tampered.tenantSelected = append(
		[]postgresv1.SelectedMigration(nil), runner.tenantSelected...,
	)
	tampered.tenantSelected[0].SHA256 = strings.Repeat("f", 64)
	transaction, err = database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if err := tampered.persistTenantRequirementCatalog(
		ctx, transaction, runnerIntegrationRunID,
	); err == nil || !strings.Contains(err.Error(), "differs") {
		t.Fatalf("tampered requirement catalog replay error = %v", err)
	}
}

func assertTenantRequirementImmutability(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	statements := []string{
		"UPDATE platform.tenant_migration_requirement_catalogs SET requirement_count=9",
		"DELETE FROM platform.tenant_migration_requirements",
	}
	for _, statement := range statements {
		if _, err := database.ExecContext(ctx, statement); err == nil {
			t.Fatalf("immutable requirement mutation passed: %s", statement)
		}
	}
}
