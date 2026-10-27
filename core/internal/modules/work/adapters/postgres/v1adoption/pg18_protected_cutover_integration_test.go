//go:build integration

package v1adoption

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestWorkPG18ProtectedCutoverAtomicReplayAndEvidence(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := startWorkPG18(t, ctx)
	prepareWorkSchemas(t, ctx, database.owner)
	authority := prepareProtectedWorkAuthority(t, ctx, database.owner, "work_adopted_a")
	if _, err := database.owner.ExecContext(ctx,
		"GRANT USAGE ON SCHEMA work_adopted_b TO harpy_adopter"); err != nil {
		t.Fatal(err)
	}
	assertProtectedWorkStageDigestGuard(t, ctx, database.owner, authority)
	assertProtectedWorkAccessDenied(t, ctx, database.owner, authority.schema)

	board := newProtectedBoardBatch(authority, "Protected A")
	applied, keyDigest, stateDigest, err := executeProtectedWorkBatch(
		ctx, database.owner, authority, authority.schema, board, false)
	if err != nil || applied != 1 || len(keyDigest) != 64 || len(stateDigest) != 64 {
		t.Fatalf("rolled back protected Work apply = %d %q %q %v",
			applied, keyDigest, stateDigest, err)
	}
	assertProtectedWorkCounts(t, ctx, database.owner, authority, 0, 0, 0, 0, 0)

	applied, keyDigest, stateDigest, err = executeProtectedWorkBatch(
		ctx, database.owner, authority, authority.schema, board, true)
	if err != nil || applied != 1 {
		t.Fatalf("protected Work apply = %d %v", applied, err)
	}
	assertProtectedWorkCounts(t, ctx, database.owner, authority, 1, 1, 1, 1, 1)
	replayed, replayKey, replayState, err := executeProtectedWorkBatch(
		ctx, database.owner, authority, authority.schema, board, true)
	if err != nil || replayed != applied || replayKey != keyDigest || replayState != stateDigest {
		t.Fatalf("protected Work exact replay = %d %q %q %v",
			replayed, replayKey, replayState, err)
	}
	assertProtectedWorkCounts(t, ctx, database.owner, authority, 1, 1, 1, 1, 1)

	tampered := newProtectedBoardBatch(authority, "Tampered A")
	if _, _, _, err = executeProtectedWorkBatch(ctx, database.owner, authority,
		authority.schema, tampered, true); err == nil {
		t.Fatal("protected Work tamper replay was accepted")
	}
	if _, _, _, err = executeProtectedWorkBatch(ctx, database.owner, authority,
		"work_adopted_b", board, true); err == nil {
		t.Fatal("protected Work cross-tenant apply was accepted")
	}
	assertProtectedTenantEmpty(t, ctx, database.owner, "work_adopted_b")
	assertProtectedWorkCounts(t, ctx, database.owner, authority, 1, 1, 1, 1, 1)

	invalidQuiet := newProtectedQuietBatch(authority, "task_board_columns", 2, "0/91")
	if _, _, _, err = executeProtectedWorkBatch(ctx, database.owner, authority,
		authority.schema, invalidQuiet, true); err == nil {
		t.Fatal("protected Work non-exact quiet barrier was accepted")
	}
	keys := []string{"task_board_columns", "board_members", "board_tasks",
		"board_task_comments", "calendar_events"}
	for index, key := range keys {
		quiet := newProtectedQuietBatch(authority, key, index+2, authority.finalLSN)
		applied, _, _, err = executeProtectedWorkBatch(
			ctx, database.owner, authority, authority.schema, quiet, true)
		if err != nil || applied != 0 {
			t.Fatalf("protected Work quiet barrier %s = %d %v", key, applied, err)
		}
	}
	assertProtectedWorkEvidence(t, ctx, database.owner, authority, "Protected A")
	assertJobReceiptCrashReplay(t, ctx, database.owner, authority.schema)
}

func assertProtectedTenantEmpty(t *testing.T, ctx context.Context,
	database *sql.DB, schema string,
) {
	t.Helper()
	workInSchema(t, ctx, database, schema, func(transaction *sql.Tx) {
		var stage, boards int
		if err := transaction.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM work_v1_adoption_stage),(SELECT count(*) FROM work_boards)`).
			Scan(&stage, &boards); err != nil {
			t.Fatal(err)
		}
		if stage != 0 || boards != 0 {
			t.Fatalf("protected Work tenant isolation stage=%d boards=%d", stage, boards)
		}
	})
}
