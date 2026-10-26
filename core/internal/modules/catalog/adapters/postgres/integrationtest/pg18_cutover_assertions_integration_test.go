//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func assertCatalogCutoverCounts(t *testing.T, ctx context.Context,
	database *sql.DB, authority catalogCutoverAuthority,
	wantStage, wantBatches, wantRuns, wantReceipts int,
) {
	t.Helper()
	var stage, batches, runs, receipts int
	err := database.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM catalog_a.catalog_v1_adoption_stage
 WHERE operation_id=$1 AND run_id=$2),
(SELECT count(*) FROM platform.cutover_batch_receipts
 WHERE operation_id=$1 AND run_id=$2),
(SELECT count(*) FROM catalog_a.catalog_adoption_runs WHERE run_id=$2),
(SELECT count(*) FROM catalog_a.catalog_v1_adoption_receipts
 WHERE operation_id=$1 AND run_id=$2)`, authority.operationID, authority.runID).
		Scan(&stage, &batches, &runs, &receipts)
	if err != nil || stage != wantStage || batches != wantBatches ||
		runs != wantRuns || receipts != wantReceipts {
		t.Fatalf("Catalog cutover counts=%d/%d/%d/%d want=%d/%d/%d/%d err=%v",
			stage, batches, runs, receipts, wantStage, wantBatches,
			wantRuns, wantReceipts, err)
	}
}

func assertCatalogCutoverEvidence(t *testing.T, ctx context.Context,
	database *sql.DB, authority catalogCutoverAuthority,
	evidence catalogCutoverEvidence,
) {
	t.Helper()
	var sourceCount, targetCount int64
	var sourceRoot, targetRoot, state string
	err := database.QueryRowContext(ctx, `SELECT source_count,target_count,
source_root_sha256,target_root_sha256,state FROM catalog_a.catalog_adoption_runs
WHERE run_id=$1`, authority.runID).Scan(&sourceCount, &targetCount,
		&sourceRoot, &targetRoot, &state)
	if err != nil || sourceCount != evidence.SourceCount ||
		targetCount != evidence.TargetCount || sourceRoot != evidence.SourceRootSHA256 ||
		targetRoot != evidence.TargetRootSHA256 || state != "VERIFIED" {
		t.Fatalf("Catalog adoption run evidence=%d/%d %q/%q %q err=%v",
			sourceCount, targetCount, sourceRoot, targetRoot, state, err)
	}
	var blocker, consumer, terminal, acceptance string
	err = database.QueryRowContext(ctx, `SELECT blocker_sha256,consumer_sha256,
evidence_sha256,acceptance_sha256 FROM catalog_a.catalog_v1_adoption_receipts
WHERE operation_id=$1 AND run_id=$2`, authority.operationID, authority.runID).
		Scan(&blocker, &consumer, &terminal, &acceptance)
	if err != nil || blocker != evidence.BlockerSHA256 ||
		consumer != evidence.ConsumerSHA256 || terminal != evidence.EvidenceSHA256 ||
		len(acceptance) != 64 {
		t.Fatalf("Catalog protected evidence=%q/%q/%q/%q err=%v",
			blocker, consumer, terminal, acceptance, err)
	}
	var installations, bindings int
	if err = database.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM catalog_a.catalog_installations),
(SELECT count(*) FROM catalog_a.catalog_component_bindings)`).
		Scan(&installations, &bindings); err != nil || installations != 1 || bindings != 1 {
		t.Fatalf("Catalog projection=%d/%d err=%v", installations, bindings, err)
	}
}

func assertCatalogCutoverDirectWritesDenied(t *testing.T, ctx context.Context,
	database *sql.DB, authority catalogCutoverAuthority,
) {
	t.Helper()
	composite := `catalog_a.catalog_v1_adoption_apply_and_record_v1(uuid,uuid,text,
uuid,smallint,text,text,pg_lsn,pg_lsn,pg_lsn,pg_lsn,xid8,integer,text,jsonb,
bytea,text,bigint,text)`
	apply := `catalog_a.catalog_v1_adoption_apply_v1(uuid,uuid,text,uuid,smallint,
text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)`
	var compositeAllowed, applyAllowed, stageAllowed, receiptAllowed bool
	err := database.QueryRowContext(ctx, `SELECT
pg_catalog.has_function_privilege('harpy_adopter',$1,'EXECUTE'),
pg_catalog.has_function_privilege('harpy_adopter',$2,'EXECUTE'),
pg_catalog.has_table_privilege('harpy_adopter',
 'catalog_a.catalog_v1_adoption_stage','DELETE'),
pg_catalog.has_table_privilege('harpy_adopter',
 'catalog_a.catalog_v1_adoption_receipts','UPDATE')`, composite, apply).
		Scan(&compositeAllowed, &applyAllowed, &stageAllowed, &receiptAllowed)
	if err != nil || !compositeAllowed || applyAllowed || stageAllowed || receiptAllowed {
		t.Fatalf("Catalog cutover ACL=%t/%t/%t/%t err=%v", compositeAllowed,
			applyAllowed, stageAllowed, receiptAllowed, err)
	}
	assertCatalogAdopterSurface(t, ctx, database)
	assertCatalogAdopterDenied(t, ctx, database,
		"DELETE FROM catalog_a.catalog_v1_adoption_stage",
		"direct DML")
	assertCatalogAdopterDenied(t, ctx, database,
		"SELECT catalog_a.catalog_v1_adoption_consumer_sha256_v1(NULL::uuid)",
		"helper invocation")
	if _, err = database.ExecContext(ctx, `UPDATE catalog_a.catalog_v1_adoption_receipts
SET acceptance_sha256=acceptance_sha256 WHERE operation_id=$1 AND run_id=$2`,
		authority.operationID, authority.runID); err == nil {
		t.Fatal("Catalog immutable receipt update was accepted")
	}
}

func assertCatalogAdopterSurface(t *testing.T, ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	var functions, tables, columns int
	var schemaCreate bool
	err := database.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM pg_catalog.pg_proc AS routine
 JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=routine.pronamespace
 WHERE namespace.nspname='catalog_a' AND pg_catalog.has_function_privilege(
   'harpy_adopter',routine.oid,'EXECUTE')),
(SELECT count(*) FROM pg_catalog.pg_class AS relation
 JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=relation.relnamespace
 WHERE namespace.nspname='catalog_a' AND relation.relkind IN ('r','p')
 AND pg_catalog.has_table_privilege('harpy_adopter',relation.oid,
   'SELECT,INSERT,UPDATE,DELETE,TRUNCATE,REFERENCES,TRIGGER')),
(SELECT count(*) FROM pg_catalog.pg_attribute AS attribute
 JOIN pg_catalog.pg_class AS relation ON relation.oid=attribute.attrelid
 JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=relation.relnamespace
 WHERE namespace.nspname='catalog_a' AND relation.relkind IN ('r','p')
 AND attribute.attnum>0 AND NOT attribute.attisdropped
 AND pg_catalog.has_column_privilege('harpy_adopter',relation.oid,
   attribute.attnum,'SELECT,INSERT,UPDATE,REFERENCES')),
pg_catalog.has_schema_privilege('harpy_adopter','catalog_a','CREATE')`).Scan(
		&functions, &tables, &columns, &schemaCreate)
	if err != nil || functions != 1 || tables != 0 || columns != 0 || schemaCreate {
		t.Fatalf("Catalog adopter surface functions/tables/columns/create=%d/%d/%d/%t err=%v",
			functions, tables, columns, schemaCreate, err)
	}
}

func assertCatalogAdopterDenied(t *testing.T, ctx context.Context,
	database *sql.DB, statement, surface string,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx,
		"SET LOCAL SESSION AUTHORIZATION harpy_adopter"); err == nil {
		_, err = transaction.ExecContext(ctx, statement)
	}
	_ = transaction.Rollback()
	if err == nil {
		t.Fatalf("Catalog adopter %s was accepted", surface)
	}
}
