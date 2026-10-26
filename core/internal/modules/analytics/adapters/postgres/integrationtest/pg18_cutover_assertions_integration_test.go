//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func assertAnalyticsCutoverCounts(t *testing.T, ctx context.Context,
	database *sql.DB, authority analyticsCutoverAuthority, stage, batches, receipts int,
) {
	t.Helper()
	var gotStage, gotBatches, gotReceipts, differentials int
	err := database.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM t_analytics_a.analytics_v1_adoption_stage
 WHERE operation_id=$1 AND run_id=$2),
(SELECT count(*) FROM platform.cutover_batch_receipts
 WHERE operation_id=$1 AND run_id=$2),
(SELECT count(*) FROM t_analytics_a.analytics_v1_adoption_receipts
 WHERE operation_id=$1 AND run_id=$2),
(SELECT count(*) FROM t_analytics_a.analytics_v1_adoption_differentials
 WHERE operation_id=$1 AND run_id=$2)`, authority.operationID, authority.runID).
		Scan(&gotStage, &gotBatches, &gotReceipts, &differentials)
	if err != nil || gotStage != stage || gotBatches != batches ||
		gotReceipts != receipts || differentials != receipts {
		t.Fatalf("Analytics cutover counts=%d/%d/%d/%d want=%d/%d/%d/%d err=%v",
			gotStage, gotBatches, gotReceipts, differentials,
			stage, batches, receipts, receipts, err)
	}
}

func assertAnalyticsCutoverProjection(t *testing.T, ctx context.Context,
	database *sql.DB, authority analyticsCutoverAuthority,
	evidence analyticsCutoverEvidence,
) {
	t.Helper()
	var queries, versions, dependencies, releases, approvals, publications, routes int
	err := database.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM t_analytics_a.analytics_saved_queries),
(SELECT count(*) FROM t_analytics_a.analytics_saved_query_versions),
(SELECT count(*) FROM t_analytics_a.analytics_query_dependencies),
(SELECT count(*) FROM t_analytics_a.analytics_publication_releases),
(SELECT count(*) FROM t_analytics_a.analytics_publication_approvals),
(SELECT count(*) FROM platform.analytics_publications WHERE tenant_id=$1),
(SELECT count(*) FROM platform.analytics_legacy_public_routes WHERE tenant_id=$1)`,
		authority.tenantID).Scan(&queries, &versions, &dependencies, &releases,
		&approvals, &publications, &routes)
	if err != nil || queries != 1 || versions != 2 || dependencies != 1 ||
		releases != 1 || approvals != 1 || publications != 1 || routes != 1 {
		t.Fatalf("Analytics projection=%d/%d/%d/%d/%d/%d/%d err=%v", queries,
			versions, dependencies, releases, approvals, publications, routes, err)
	}
	var source, target, dependency, public, zero, rollback, watermark, terminal string
	err = database.QueryRowContext(ctx, `SELECT source_root_sha256,target_root_sha256,
dependency_sha256,public_approval_sha256,zero_hit_sha256,rollback_sha256,
watermark_sha256,evidence_sha256 FROM t_analytics_a.analytics_v1_adoption_receipts
WHERE operation_id=$1 AND run_id=$2`, authority.operationID, authority.runID).
		Scan(&source, &target, &dependency, &public, &zero, &rollback, &watermark, &terminal)
	if err != nil || source != evidence.SourceRootSHA256 || target != evidence.TargetRootSHA256 ||
		dependency != evidence.DependencySHA256 || public != evidence.PublicApprovalSHA256 ||
		zero != evidence.ZeroHitSHA256 || rollback != evidence.RollbackSHA256 ||
		watermark != evidence.WatermarkSHA256 || terminal != evidence.EvidenceSHA256 {
		t.Fatalf("Analytics protected evidence differs: %v", err)
	}
	var batchOrdinal, appliedChanges int
	var commitLSN string
	err = database.QueryRowContext(ctx, `SELECT batch_ordinal,commit_lsn::text,
(SELECT count(*) FROM platform.applied_change_receipts
 WHERE operation_id=$1 AND run_id=$2)
FROM t_analytics_a.analytics_v1_adoption_stage
WHERE operation_id=$1 AND run_id=$2`, authority.operationID, authority.runID).
		Scan(&batchOrdinal, &commitLSN, &appliedChanges)
	if err != nil || batchOrdinal != 2 || commitLSN != "0/50" || appliedChanges != 2 {
		t.Fatalf("Analytics incremental watermark=%d/%q/%d err=%v",
			batchOrdinal, commitLSN, appliedChanges, err)
	}
}

func assertAnalyticsCutoverACL(t *testing.T, ctx context.Context,
	database *sql.DB, authority analyticsCutoverAuthority,
) {
	t.Helper()
	var functions, tables, columns int
	var create bool
	err := database.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM pg_proc AS routine JOIN pg_namespace AS namespace
 ON namespace.oid=routine.pronamespace WHERE namespace.nspname='t_analytics_a'
 AND has_function_privilege('harpy_adopter',routine.oid,'EXECUTE')),
(SELECT count(*) FROM pg_class AS relation JOIN pg_namespace AS namespace
 ON namespace.oid=relation.relnamespace WHERE namespace.nspname='t_analytics_a'
 AND relation.relkind IN('r','p') AND has_table_privilege('harpy_adopter',
 relation.oid,'SELECT,INSERT,UPDATE,DELETE,TRUNCATE,REFERENCES,TRIGGER')),
(SELECT count(*) FROM pg_attribute AS attribute JOIN pg_class AS relation
 ON relation.oid=attribute.attrelid JOIN pg_namespace AS namespace
 ON namespace.oid=relation.relnamespace WHERE namespace.nspname='t_analytics_a'
 AND relation.relkind IN('r','p') AND attribute.attnum>0 AND NOT attribute.attisdropped
 AND has_column_privilege('harpy_adopter',relation.oid,attribute.attnum,
 'SELECT,INSERT,UPDATE,REFERENCES')),
has_schema_privilege('harpy_adopter','t_analytics_a','CREATE')`).Scan(
		&functions, &tables, &columns, &create)
	if err != nil || functions != 1 || tables != 0 || columns != 0 || create {
		t.Fatalf("Analytics adopter surface=%d/%d/%d/%t err=%v",
			functions, tables, columns, create, err)
	}
	assertAnalyticsAdopterDenied(t, ctx, database,
		"SELECT t_analytics_a.analytics_v1_adoption_dependency_sha256_v1('"+
			authority.operationID+"','"+authority.runID+"','"+authority.tenantID+"')")
	assertAnalyticsAdopterDenied(t, ctx, database,
		"SELECT * FROM t_analytics_a.analytics_adoption_accept_v1(NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL)")
	if _, err = database.ExecContext(ctx, `UPDATE t_analytics_a.analytics_v1_adoption_receipts
SET acceptance_sha256=acceptance_sha256 WHERE operation_id=$1 AND run_id=$2`,
		authority.operationID, authority.runID); err == nil {
		t.Fatal("Analytics immutable receipt update was accepted")
	}
}

func assertAnalyticsAdopterDenied(t *testing.T, ctx context.Context,
	database *sql.DB, statement string,
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
		t.Fatalf("Analytics adopter statement was accepted: %s", statement)
	}
}
