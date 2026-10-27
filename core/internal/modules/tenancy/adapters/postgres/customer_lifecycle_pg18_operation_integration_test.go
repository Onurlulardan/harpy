//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"
)

type lifecycleOperationResult struct {
	outcome, state string
	id             sql.NullString
	revision       sql.NullInt64
	replay         sql.NullBool
	err            error
}

func assertLifecycleOffboardAndPurge(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	offboard := acceptLifecycleOperation(ctx, database, lifecyclePurgeID, "offboard", 1,
		"tenant-lifecycle-offboard-1", bytes.Repeat([]byte{0x41}, 32), false)
	if offboard.err != nil || offboard.outcome != "accepted" || !offboard.id.Valid ||
		offboard.state != "REQUESTED" {
		t.Fatalf("offboard accept=%+v", offboard)
	}
	replay := acceptLifecycleOperation(ctx, database, lifecyclePurgeID, "offboard", 1,
		"tenant-lifecycle-offboard-1", bytes.Repeat([]byte{0x41}, 32), false)
	if replay.err != nil || replay.outcome != "replay" || !replay.replay.Bool ||
		replay.id.String != offboard.id.String {
		t.Fatalf("offboard replay=%+v", replay)
	}
	claimID, kind, schema := claimLifecycleOperation(t, ctx, database)
	if claimID != offboard.id.String || kind != "offboard" || schema == "" {
		t.Fatalf("offboard claim id=%s kind=%s schema=%s", claimID, kind, schema)
	}
	state, revision := completeLifecycleOperation(t, ctx, database, claimID, false)
	if state != "SUCCEEDED" || revision != 2 {
		t.Fatalf("offboard complete state=%s revision=%d", state, revision)
	}
	assertLifecycleTenantTimes(t, ctx, database, lifecyclePurgeID, true, false, 2)

	assertPurgeWithoutTrustedProofRejected(t, ctx, database)
	purge := acceptLifecycleOperation(ctx, database, lifecyclePurgeID, "purge", 2,
		"tenant-lifecycle-purge-0001", bytes.Repeat([]byte{0x42}, 32), true)
	if purge.err != nil || purge.outcome != "accepted" || !purge.id.Valid {
		t.Fatalf("purge accept=%+v", purge)
	}
	claimID, kind, schema = claimLifecycleOperation(t, ctx, database)
	if claimID != purge.id.String || kind != "purge" || schema == "" {
		t.Fatalf("purge claim id=%s kind=%s schema=%s", claimID, kind, schema)
	}
	state, revision = completeLifecycleOperation(t, ctx, database, claimID, true)
	if state != "SUCCEEDED" || revision != 3 {
		t.Fatalf("purge complete state=%s revision=%d", state, revision)
	}
	assertLifecycleTenantTimes(t, ctx, database, lifecyclePurgeID, true, true, 3)
}

func acceptLifecycleOperation(ctx context.Context, database *sql.DB, targetID, kind string,
	expected int64, key string, digest []byte, proof bool,
) lifecycleOperationResult {
	var result lifecycleOperationResult
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		result.err = err
		return result
	}
	if _, err = transaction.ExecContext(ctx, `SET LOCAL SESSION AUTHORIZATION harpy_v2_runtime;
SET LOCAL ROLE harpy_runtime`); err != nil {
		_ = transaction.Rollback()
		result.err = err
		return result
	}
	arguments := []any{testEnvironmentID, lifecycleControlID, lifecycleActorID,
		targetID, kind, expected, key, digest, nil, nil, nil, nil, nil,
		time.Now().UTC().Truncate(time.Microsecond)}
	if proof {
		arguments[8] = "80000000-0000-4000-8000-000000000001"
		arguments[9] = "80000000-0000-4000-8000-000000000002"
		arguments[10] = "80000000-0000-4000-8000-000000000003"
		arguments[11] = bytes.Repeat([]byte{0x51}, 32)
		arguments[12] = time.Now().UTC().Add(-time.Minute).Truncate(time.Microsecond)
	}
	var control, target, operationKind sql.NullString
	var created, updated, completed sql.NullTime
	var failure sql.NullString
	result.err = transaction.QueryRowContext(ctx, `SELECT *
FROM platform.tenancy_customer_lifecycle_accept_v1(
$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`, arguments...).Scan(
		&result.outcome, &result.id, &control, &target, &operationKind, &result.state,
		&result.revision, &result.replay, &created, &updated, &completed, &failure)
	if result.err == nil {
		result.err = transaction.Commit()
	} else {
		_ = transaction.Rollback()
	}
	return result
}

func claimLifecycleOperation(t *testing.T, ctx context.Context, database *sql.DB) (
	string, string, string,
) {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	var operationID, control, target, kind, state, schema string
	var revision int64
	var created, updated time.Time
	var completed sql.NullTime
	var failure sql.NullString
	var retention, backup, contributor sql.NullString
	var evidence []byte
	var verified sql.NullTime
	err := transaction.QueryRowContext(ctx, `SELECT *
FROM platform.tenancy_customer_lifecycle_claim_v1($1,$2,120)`,
		lifecycleWorkerID, time.Now().UTC().Truncate(time.Microsecond)).Scan(
		&operationID, &control, &target, &kind, &state, &revision, &created, &updated,
		&completed, &failure, &schema, &retention, &backup, &contributor, &evidence, &verified)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return operationID, kind, schema
}

func completeLifecycleOperation(t *testing.T, ctx context.Context, database *sql.DB,
	operationID string, purge bool,
) (string, int64) {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	statement := `SELECT * FROM platform.tenancy_customer_lifecycle_complete_offboard_v1($1,$2,$3)`
	arguments := []any{operationID, lifecycleWorkerID, time.Now().UTC().Truncate(time.Microsecond)}
	if purge {
		statement = `SELECT * FROM platform.tenancy_customer_lifecycle_complete_purge_v1($1,$2,$3,$4,$5)`
		arguments = []any{operationID, lifecycleWorkerID,
			"90000000-0000-4000-8000-000000000001", bytes.Repeat([]byte{0x61}, 32),
			time.Now().UTC().Truncate(time.Microsecond)}
	}
	var id, control, target, kind, state string
	var revision int64
	var created, updated time.Time
	var completed sql.NullTime
	var failure sql.NullString
	err := transaction.QueryRowContext(ctx, statement, arguments...).Scan(&id, &control,
		&target, &kind, &state, &revision, &created, &updated, &completed, &failure)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return state, revision
}
