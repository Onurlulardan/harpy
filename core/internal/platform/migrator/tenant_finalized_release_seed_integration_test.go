//go:build integration

package migrator

import (
	"context"
	"os"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func seedFinalizedRelease(
	t *testing.T, ctx context.Context, fixture tenantMigrationFixture,
	runner *Runner, selected postgresv1.SelectedMigration,
) {
	t.Helper()
	historySQL, err := os.ReadFile("adapters/postgres/sql/create_history_table.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.database.ExecContext(ctx, string(historySQL)); err != nil {
		t.Fatal(err)
	}
	identity := runner.identity
	if _, err := fixture.database.ExecContext(ctx, `INSERT INTO platform.environment_identity
(environment_id,database_origin,database_system_identifier,database_oid,
release_id,bootstrap_authority_sha256) VALUES($1,$2,$3,$4,$5,$6)`,
		identity.environmentID, identity.databaseOrigin, identity.databaseSystemID,
		identity.databaseOID, identity.releaseID, identity.authoritySHA256); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.database.ExecContext(ctx, `INSERT INTO platform.database_generation_manifests
(environment_id,database_generation,migration_manifest_sha256,extension_manifest_sha256)
VALUES($1,$2,$3,$4)`, identity.environmentID, identity.databaseGeneration,
		identity.manifestSHA256, identity.extensionSHA256); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.database.ExecContext(ctx, `INSERT INTO platform.platform_migration_history
(version,record_kind,is_applied,owner,scope,origin,selected_path,selected_sha256,
run_id,target_database_generation,next_minimum_runtime_generation,applied_at)
VALUES($1,'migration',true,$2,$3,$4,$5,$6,$7,$8,$9,pg_catalog.clock_timestamp())`,
		selected.Version, selected.Owner, selected.Scope, selected.Origin, selected.Path,
		selected.SHA256, finalizedMigrationRun, selected.TargetDatabaseGeneration,
		selected.NextMinimumRuntimeGeneration); err != nil {
		t.Fatal(err)
	}
	transaction, err := fixture.database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err := transaction.ExecContext(ctx, `INSERT INTO platform.migration_runs
(run_id,release_id,environment_id,database_origin,command,state,
migration_manifest_sha256,extension_manifest_sha256,roles_manifest_sha256,
privileges_manifest_sha256,ownership_manifest_sha256,started_at,finished_at)
VALUES($1,$2,$3,$4,'up','succeeded',$5,$6,$7,$7,$7,
pg_catalog.statement_timestamp(),pg_catalog.statement_timestamp())`,
		finalizedMigrationRun, identity.releaseID, identity.environmentID,
		identity.databaseOrigin, identity.manifestSHA256, identity.extensionSHA256,
		strings.Repeat("d", 64)); err != nil {
		t.Fatal(err)
	}
	for index, receipt := range runner.evidence.receipts {
		digit := string(rune('1' + index))
		if _, err := transaction.ExecContext(ctx, `INSERT INTO platform.bootstrap_receipts
(kind,run_id,environment_id,release_id,database_origin,database_system_identifier,
database_oid,operation_digest,canonical_document,canonical_sha256,signature_key_id,issued_at)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,'{}',$9,'test_key',pg_catalog.statement_timestamp())`,
			receipt.kind, receipt.runID, identity.environmentID, identity.releaseID,
			identity.databaseOrigin, identity.databaseSystemID, identity.databaseOID,
			strings.Repeat(digit, 64), strings.Repeat(string(rune('4'+index)), 64)); err != nil {
			t.Fatal(err)
		}
	}
	contract := runner.schemaAuthority.Contract()
	if _, err := transaction.ExecContext(ctx, `INSERT INTO platform.schema_contracts
(component,database_generation,minimum_runtime_generation,migration_manifest_sha256,
extension_manifest_sha256,status,verified_run_id,revision,verified_at)
VALUES('core',$1,$2,$3,$4,'ready',$5,1,pg_catalog.statement_timestamp())`,
		identity.databaseGeneration, identity.minimumRuntimeGeneration,
		identity.manifestSHA256, identity.extensionSHA256, finalizedMigrationRun); err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.ExecContext(ctx, `INSERT INTO platform.release_generations
(migration_run_id,release_id,environment_id,database_origin,database_generation,
minimum_runtime_generation,runtime_generation,minimum_database_generation,
migration_manifest_sha256,extension_manifest_sha256,postgres_image_digest,
bootstrap_receipt_run_id,extension_receipt_run_id,cleanup_receipt_run_id)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		finalizedMigrationRun, identity.releaseID, identity.environmentID,
		identity.databaseOrigin, identity.databaseGeneration, identity.minimumRuntimeGeneration,
		contract.RuntimeGeneration, contract.MinimumDatabaseGeneration,
		identity.manifestSHA256, identity.extensionSHA256, contract.PostgresImageDigest,
		runner.evidence.receipts[0].runID, runner.evidence.receipts[1].runID,
		runner.evidence.receipts[2].runID); err != nil {
		t.Fatal(err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}
