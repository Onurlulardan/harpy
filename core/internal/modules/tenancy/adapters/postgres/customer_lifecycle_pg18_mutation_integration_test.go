//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"
)

type lifecycleMutationResult struct {
	outcome          string
	replay           bool
	id, name, status sql.NullString
	revision         sql.NullInt64
	err              error
}

func assertLifecycleMutations(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	digest := bytes.Repeat([]byte{0x31}, 32)
	update := executeLifecycleMutation(ctx, database, lifecycleTargetID, "update",
		"Renamed customer", 1, "tenant-lifecycle-update-01", digest)
	if update.err != nil || update.outcome != "applied" || update.replay ||
		!update.revision.Valid || update.revision.Int64 != 2 || update.name.String != "Renamed customer" {
		t.Fatalf("lifecycle update=%+v", update)
	}
	replay := executeLifecycleMutation(ctx, database, lifecycleTargetID, "update",
		"Renamed customer", 1, "tenant-lifecycle-update-01", digest)
	if replay.err != nil || replay.outcome != "replay" || !replay.replay ||
		replay.revision.Int64 != 2 {
		t.Fatalf("lifecycle replay=%+v", replay)
	}
	conflict := executeLifecycleMutation(ctx, database, lifecycleTargetID, "update",
		"Other name", 1, "tenant-lifecycle-update-01", bytes.Repeat([]byte{0x32}, 32))
	if conflict.err != nil || conflict.outcome != "idempotency_conflict" || conflict.id.Valid {
		t.Fatalf("lifecycle idempotency conflict=%+v", conflict)
	}
	suspended := executeLifecycleMutation(ctx, database, lifecycleTargetID, "suspend",
		"", 2, "tenant-lifecycle-suspend-1", bytes.Repeat([]byte{0x33}, 32))
	if suspended.err != nil || suspended.outcome != "applied" ||
		suspended.status.String != "SUSPENDED" || suspended.revision.Int64 != 3 {
		t.Fatalf("lifecycle suspend=%+v", suspended)
	}
	reactivated := executeLifecycleMutation(ctx, database, lifecycleTargetID, "reactivate",
		"", 3, "tenant-lifecycle-react-01", bytes.Repeat([]byte{0x34}, 32))
	if reactivated.err != nil || reactivated.outcome != "applied" ||
		reactivated.status.String != "ACTIVE" || reactivated.revision.Int64 != 4 {
		t.Fatalf("lifecycle reactivate=%+v", reactivated)
	}
	stale := executeLifecycleMutation(ctx, database, lifecycleTargetID, "suspend",
		"", 3, "tenant-lifecycle-stale-001", bytes.Repeat([]byte{0x35}, 32))
	if stale.err != nil || stale.outcome != "revision_conflict" || stale.id.Valid {
		t.Fatalf("lifecycle stale revision=%+v", stale)
	}
}

func executeLifecycleMutation(ctx context.Context, database *sql.DB, targetID, kind,
	name string, expected int64, key string, digest []byte,
) lifecycleMutationResult {
	var result lifecycleMutationResult
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
	var nullableName any
	if name != "" {
		nullableName = name
	}
	var slug sql.NullString
	var settingsRevision sql.NullInt64
	var createdAt, updatedAt, offboardedAt, purgedAt sql.NullTime
	result.err = transaction.QueryRowContext(ctx, `SELECT *
FROM platform.tenancy_customer_lifecycle_mutate_v1(
$1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, testEnvironmentID, lifecycleControlID,
		lifecycleActorID, targetID, kind, nullableName, expected, key, digest,
		time.Now().UTC().Truncate(time.Microsecond)).Scan(&result.outcome, &result.replay,
		&result.id, &result.name, &slug, &result.status, &result.revision,
		&settingsRevision, &createdAt, &updatedAt, &offboardedAt, &purgedAt)
	if result.err == nil {
		result.err = transaction.Commit()
	} else {
		_ = transaction.Rollback()
	}
	return result
}
