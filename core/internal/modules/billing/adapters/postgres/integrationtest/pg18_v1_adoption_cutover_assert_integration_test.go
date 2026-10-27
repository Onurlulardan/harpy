//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func assertProtectedBillingAccessDenied(t *testing.T, ctx context.Context,
	database *sql.DB, operationID, runID any,
) {
	t.Helper()
	statements := []struct {
		query string
		args  []any
	}{
		{"SET LOCAL ROLE harpy_owner", nil},
		{"UPDATE platform.billing_v1_adoption_plan_chunks SET created_at=created_at WHERE false", nil},
		{"UPDATE platform.billing_v1_adoption_runs SET updated_at=updated_at WHERE false", nil},
		{"SELECT platform.billing_convert_v1_storage_megabytes_v1(NULL::bigint)", nil},
		{`SELECT platform.billing_register_v1_adoption_v1(
NULL::uuid,NULL::uuid,NULL::uuid,NULL::bytea,NULL::text,NULL::bytea,NULL::text)`, nil},
		{"SELECT platform.billing_apply_v1_adoption_v1($1,$2)", []any{operationID, runID}},
		{"SELECT * FROM platform.billing_verify_v1_adoption_v1($1,$2)", []any{operationID, runID}},
		{`SELECT * FROM platform.billing_rehearse_v1_adoption_rollback_v1(
NULL::uuid,NULL::uuid,NULL::uuid,NULL::text,NULL::timestamptz)`, nil},
		{`SELECT platform.billing_assert_v1_adoption_ready_v1(
NULL::uuid,NULL::uuid,NULL::text)`, nil},
	}
	for _, statement := range statements {
		transaction, err := database.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, err = transaction.ExecContext(ctx,
			"SET LOCAL SESSION AUTHORIZATION harpy_adopter")
		if err == nil {
			_, err = transaction.ExecContext(ctx, statement.query, statement.args...)
		}
		_ = transaction.Rollback()
		if err == nil {
			t.Fatalf("Billing adopter direct authority allowed: %s", statement.query)
		}
	}
}

func assertProtectedBillingCounts(t *testing.T, ctx context.Context,
	database *sql.DB, a protectedBillingAuthority,
	chunks, runs, batches, checkpoints, receipts int,
) {
	t.Helper()
	var actualChunks, actualRuns, actualBatches, actualCheckpoints, actualReceipts int
	err := database.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM platform.billing_v1_adoption_plan_chunks
 WHERE operation_id=$1 AND run_id=$2),
(SELECT count(*) FROM platform.billing_v1_adoption_runs
 WHERE operation_id=$1 AND run_id=$2),
(SELECT count(*) FROM platform.cutover_batch_receipts
 WHERE operation_id=$1 AND run_id=$2),
(SELECT count(*) FROM platform.cutover_copy_checkpoints
 WHERE operation_id=$1 AND run_id=$2),
(SELECT count(*) FROM platform.applied_change_receipts
 WHERE operation_id=$1 AND run_id=$2)`, a.operationID, a.runID).Scan(
		&actualChunks, &actualRuns, &actualBatches, &actualCheckpoints, &actualReceipts)
	if err != nil {
		t.Fatal(err)
	}
	if actualChunks != chunks || actualRuns != runs || actualBatches != batches ||
		actualCheckpoints != checkpoints || actualReceipts != receipts {
		t.Fatalf("Billing protected counts chunks=%d runs=%d batches=%d checkpoints=%d receipts=%d",
			actualChunks, actualRuns, actualBatches, actualCheckpoints, actualReceipts)
	}
}

func assertProtectedBillingRun(t *testing.T, ctx context.Context,
	database *sql.DB, a protectedBillingAuthority, expectedState string,
) {
	t.Helper()
	var state, planSHA string
	var money, period, coupon, usage sql.NullInt64
	var sourcePresent, baselineTargetEqual sql.NullBool
	err := database.QueryRowContext(ctx, `SELECT state,encode(plan_sha256,'hex'),
money_difference,period_difference,coupon_difference,usage_difference,
source_sha256 IS NOT NULL,baseline_sha256=target_sha256
FROM platform.billing_v1_adoption_runs
WHERE operation_id=$1 AND run_id=$2`, a.operationID, a.runID).Scan(
		&state, &planSHA, &money, &period, &coupon, &usage,
		&sourcePresent, &baselineTargetEqual)
	if err != nil || state != expectedState || planSHA != a.planSHA {
		t.Fatalf("Billing protected run state=%q plan=%q err=%v", state, planSHA, err)
	}
	if expectedState != "rollback_rehearsed" {
		return
	}
	for name, difference := range map[string]sql.NullInt64{
		"money": money, "period": period, "coupon": coupon, "usage": usage,
	} {
		if !difference.Valid || difference.Int64 != 0 {
			t.Fatalf("Billing %s difference = %#v", name, difference)
		}
	}
	if !sourcePresent.Valid || !sourcePresent.Bool ||
		!baselineTargetEqual.Valid || !baselineTargetEqual.Bool {
		t.Fatal("Billing protected reconciliation digests differ")
	}
}

func assertProtectedBillingIsolationEmpty(t *testing.T, ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	var chunks, runs int
	err := database.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM platform.billing_v1_adoption_plan_chunks WHERE scope_id=$1),
(SELECT count(*) FROM platform.billing_v1_adoption_runs WHERE scope_id=$1)`,
		billingIsolationTenantID).Scan(&chunks, &runs)
	if err != nil || chunks != 0 || runs != 0 {
		t.Fatalf("Billing isolated tenant chunks=%d runs=%d err=%v", chunks, runs, err)
	}
}

func assertProtectedBillingEvidence(t *testing.T, ctx context.Context,
	database *sql.DB, global, tenant protectedBillingAuthority,
) {
	t.Helper()
	var groupOrder string
	if err := database.QueryRowContext(ctx, `SELECT string_agg(domain,',' ORDER BY ordinal)
FROM platform.cutover_group_members
WHERE scope_kind='tenant' AND scope_id=$1 AND group_id='billing'`,
		tenant.scopeID).Scan(&groupOrder); err != nil || groupOrder != "billing" {
		t.Fatalf("Billing tenant cutover group order=%q err=%v", groupOrder, err)
	}
	var globalBatches, tenantBatches, complete, receipts, rollbackReceipts int
	err := database.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM platform.cutover_batch_receipts WHERE operation_id=$1 AND run_id=$2),
(SELECT count(*) FROM platform.cutover_batch_receipts WHERE operation_id=$3 AND run_id=$4),
(SELECT count(*) FROM platform.cutover_copy_checkpoints
 WHERE state='complete' AND ((operation_id=$1 AND run_id=$2) OR
 (operation_id=$3 AND run_id=$4))),
(SELECT count(*) FROM platform.applied_change_receipts
 WHERE (operation_id=$1 AND run_id=$2) OR (operation_id=$3 AND run_id=$4)),
(SELECT count(*) FROM platform.billing_v1_adoption_rollback_rehearsals
 WHERE (operation_id=$1 AND run_id=$2) OR (operation_id=$3 AND run_id=$4))`,
		global.operationID, global.runID, tenant.operationID, tenant.runID).Scan(
		&globalBatches, &tenantBatches, &complete, &receipts, &rollbackReceipts)
	if err != nil || globalBatches != 2 || tenantBatches != 1 || complete != 2 ||
		receipts != 2 || rollbackReceipts != 2 {
		t.Fatalf("Billing evidence global=%d tenant=%d complete=%d receipts=%d rollback=%d err=%v",
			globalBatches, tenantBatches, complete, receipts, rollbackReceipts, err)
	}
	var xidDifference, invalidReceipt int
	err = database.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM (
 (SELECT operation_id,run_id,source_xid FROM platform.cutover_change_journal
  EXCEPT SELECT operation_id,run_id,source_xid FROM platform.applied_change_receipts)
 UNION ALL
 (SELECT operation_id,run_id,source_xid FROM platform.applied_change_receipts
  EXCEPT SELECT operation_id,run_id,source_xid FROM platform.cutover_change_journal)
) AS difference),
(SELECT count(*) FROM platform.applied_change_receipts
 WHERE relation_ordinal<>1 OR change_ordinal<>1 OR mutation_kind<>'insert')`).Scan(
		&xidDifference, &invalidReceipt)
	if err != nil || xidDifference != 0 || invalidReceipt != 0 {
		t.Fatalf("Billing receipt evidence xid_diff=%d invalid=%d err=%v",
			xidDifference, invalidReceipt, err)
	}
}
