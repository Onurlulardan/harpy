//go:build integration

package v1adoption

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

func assertProtectedWorkStageDigestGuard(t *testing.T, ctx context.Context,
	database *sql.DB, a protectedWorkAuthority,
) {
	t.Helper()
	transaction, err := beginWorkSchema(ctx, database, a.schema)
	if err != nil {
		t.Fatal(err)
	}
	_, err = transaction.ExecContext(ctx, `INSERT INTO work_v1_adoption_stage
(operation_id,run_id,source_kind,source_id,target_kind,target_id,source_sha256,
 target_sha256,target_record,removed,source_xid,commit_lsn,batch_ordinal,
 change_ordinal,updated_at)
VALUES ($1,$2,'task_boards','41000000-0000-4000-8000-000000000201','BOARD',
 '41000000-0000-4000-8000-000000000202',decode(repeat('01',32),'hex'),
 decode(repeat('00',32),'hex'),'{}'::jsonb,false,'4001'::xid8,'0/30',1,1,
 pg_catalog.clock_timestamp())`, a.operationID, a.runID)
	_ = transaction.Rollback()
	if err == nil {
		t.Fatal("Work stage accepted a target record with a mismatched digest")
	}
}

func assertProtectedWorkAccessDenied(t *testing.T, ctx context.Context,
	database *sql.DB, schema string,
) {
	t.Helper()
	statements := []string{
		"UPDATE work_v1_adoption_stage SET updated_at=updated_at WHERE false",
		"CREATE TABLE work_cutover_forbidden(id integer)",
		"UPDATE platform.cutover_operations SET updated_at=updated_at WHERE false",
		`SELECT * FROM work_v1_adoption_apply_v1(
NULL::uuid,NULL::uuid,NULL::text,NULL::uuid,NULL::smallint,NULL::text,
NULL::pg_lsn,NULL::pg_lsn,NULL::pg_lsn,NULL::xid8,NULL::integer,NULL::jsonb,NULL::text)`,
	}
	for _, statement := range statements {
		transaction, err := database.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, err = transaction.ExecContext(ctx,
			"SET LOCAL SESSION AUTHORIZATION harpy_adopter")
		if err == nil {
			_, err = transaction.ExecContext(ctx,
				"SELECT pg_catalog.set_config('search_path',$1,true)", schema+",pg_catalog")
		}
		if err == nil {
			_, err = transaction.ExecContext(ctx, statement)
		}
		_ = transaction.Rollback()
		if err == nil {
			t.Fatalf("adopter direct authority unexpectedly allowed: %s", statement)
		}
	}
}

func assertProtectedWorkCounts(t *testing.T, ctx context.Context, database *sql.DB,
	a protectedWorkAuthority, stage, runs, batches, checkpoints, receipts int,
) {
	t.Helper()
	workInSchema(t, ctx, database, a.schema, func(transaction *sql.Tx) {
		var actualStage, actualRuns int
		if err := transaction.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM work_v1_adoption_stage WHERE operation_id=$1 AND run_id=$2),
(SELECT count(*) FROM work_adoption_runs WHERE run_id=$2)`,
			a.operationID, a.runID).Scan(&actualStage, &actualRuns); err != nil {
			t.Fatal(err)
		}
		if actualStage != stage || actualRuns != runs {
			t.Fatalf("Work protected local counts stage=%d runs=%d", actualStage, actualRuns)
		}
	})
	var actualBatches, actualCheckpoints, actualReceipts int
	if err := database.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM platform.cutover_batch_receipts WHERE operation_id=$1 AND run_id=$2),
(SELECT count(*) FROM platform.cutover_copy_checkpoints WHERE operation_id=$1 AND run_id=$2),
(SELECT count(*) FROM platform.applied_change_receipts WHERE operation_id=$1 AND run_id=$2)`,
		a.operationID, a.runID).Scan(&actualBatches, &actualCheckpoints, &actualReceipts); err != nil {
		t.Fatal(err)
	}
	if actualBatches != batches || actualCheckpoints != checkpoints || actualReceipts != receipts {
		t.Fatalf("Work protected platform counts batches=%d checkpoints=%d receipts=%d",
			actualBatches, actualCheckpoints, actualReceipts)
	}
}

func assertProtectedWorkEvidence(t *testing.T, ctx context.Context, database *sql.DB,
	a protectedWorkAuthority, expectedName string,
) {
	t.Helper()
	workInSchema(t, ctx, database, a.schema, func(transaction *sql.Tx) {
		var boardName, runState, fenceState, sourceDigest, targetDigest string
		var sourceCount, targetCount int
		if err := transaction.QueryRowContext(ctx, `SELECT board.name,run.state,
fence.state,run.source_count,run.target_count,encode(run.source_sha256,'hex'),
encode(run.target_sha256,'hex')
FROM work_boards AS board CROSS JOIN work_adoption_runs AS run
CROSS JOIN work_v1_writer_fence AS fence
WHERE run.run_id=$1 AND fence.singleton`, a.runID).Scan(&boardName, &runState,
			&fenceState, &sourceCount, &targetCount, &sourceDigest, &targetDigest); err != nil {
			t.Fatal(err)
		}
		if boardName != expectedName || runState != "ACCEPTED" || fenceState != "V2_FENCED" ||
			sourceCount != 1 || targetCount != 1 || len(sourceDigest) != 64 || len(targetDigest) != 64 {
			t.Fatalf("Work protected evidence board=%s run=%s fence=%s source=%d target=%d",
				boardName, runState, fenceState, sourceCount, targetCount)
		}
		var digestMatch bool
		if err := transaction.QueryRowContext(ctx, `SELECT
encode(run.source_sha256,'hex')=pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
 stage.source_kind||':'||stage.source_id::text||':'||encode(stage.source_sha256,'hex'),'UTF8')),'hex')
AND encode(run.target_sha256,'hex')=pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
 stage.target_kind||':'||stage.target_id::text||':'||encode(stage.target_sha256,'hex'),'UTF8')),'hex')
FROM work_adoption_runs AS run CROSS JOIN work_v1_adoption_stage AS stage
WHERE run.run_id=$1 AND stage.operation_id=$2 AND stage.run_id=$1`,
			a.runID, a.operationID).Scan(&digestMatch); err != nil || !digestMatch {
			t.Fatalf("Work protected source/target digest evidence differs: %v", err)
		}
	})
	assertProtectedPlatformEvidence(t, ctx, database, a)
}

func assertProtectedPlatformEvidence(t *testing.T, ctx context.Context,
	database *sql.DB, a protectedWorkAuthority,
) {
	t.Helper()
	var batches, complete, applied, quiet, xidDifference int
	err := database.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM platform.cutover_batch_receipts WHERE operation_id=$1 AND run_id=$2),
(SELECT count(*) FROM platform.cutover_copy_checkpoints WHERE operation_id=$1 AND run_id=$2
 AND state='complete' AND final_lsn=$3::pg_lsn),
(SELECT count(*) FROM platform.applied_change_receipts WHERE operation_id=$1 AND run_id=$2
 AND manifest_sha256=$4 AND relation_ordinal=1 AND change_ordinal=1
 AND mutation_kind='insert'),
(SELECT count(*) FROM platform.cutover_batch_receipts AS batch
 JOIN platform.cutover_copy_checkpoints AS checkpoint
 USING (operation_id,run_id,relation_ordinal,checkpoint_key)
 WHERE batch.operation_id=$1 AND batch.run_id=$2 AND batch.applied_count=0
 AND batch.source_commit_lsn=checkpoint.final_lsn),
(SELECT count(*) FROM (
 (SELECT source_xid FROM platform.cutover_change_journal WHERE operation_id=$1 AND run_id=$2
  EXCEPT SELECT source_xid FROM platform.applied_change_receipts WHERE operation_id=$1 AND run_id=$2)
 UNION ALL
 (SELECT source_xid FROM platform.applied_change_receipts WHERE operation_id=$1 AND run_id=$2
  EXCEPT SELECT source_xid FROM platform.cutover_change_journal WHERE operation_id=$1 AND run_id=$2)
) AS difference)`, a.operationID, a.runID, a.finalLSN, a.leaseManifest).
		Scan(&batches, &complete, &applied, &quiet, &xidDifference)
	if err != nil {
		t.Fatal(err)
	}
	if batches != 6 || complete != 6 || applied != 1 || quiet != 5 || xidDifference != 0 {
		t.Fatalf("Work platform evidence batches=%d complete=%d applied=%d quiet=%d xid_diff=%d",
			batches, complete, applied, quiet, xidDifference)
	}
	var ordinals string
	if err = database.QueryRowContext(ctx, `SELECT string_agg(batch_ordinal::text,',' ORDER BY batch_ordinal)
FROM platform.cutover_batch_receipts WHERE operation_id=$1 AND run_id=$2`,
		a.operationID, a.runID).Scan(&ordinals); err != nil || strings.TrimSpace(ordinals) != "1,2,3,4,5,6" {
		t.Fatalf("Work global batch ordinals = %q, %v", ordinals, err)
	}
}
