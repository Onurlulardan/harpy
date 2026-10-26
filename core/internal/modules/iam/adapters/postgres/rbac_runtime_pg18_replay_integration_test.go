//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"strings"
	"testing"
)

const (
	iamRBACRuntimeActionID    = "98000000-0000-4000-8000-000000000001"
	iamRBACRuntimeOperationID = "98100000-0000-4000-8000-000000000001"
)

const createIAMRBACRuntimeAction = `
SELECT pg_catalog.concat_ws('|', result_id::text, result_revision::text,
    authorization_clock::text, replayed::text)
FROM iam_rbac_action_create_v1(
    pg_catalog.current_schema()::name, $1::uuid, $2::bytea, $3::uuid,
    $4::text, 'Runtime Create', 'runtime test', false,
    '51000000-0000-4000-8000-000000000001',
    '93000000-0000-4000-8000-000000000001', $5::bigint, $6::bigint
)`

func assertIAMRBACRuntimeReplayAndIsolation(t *testing.T, ctx context.Context,
	database *sql.DB) {
	t.Helper()
	hash := bytes.Repeat([]byte{0x41}, 32)
	arguments := []any{iamRBACRuntimeOperationID, hash, iamRBACRuntimeActionID,
		"runtime_create", int64(1), int64(1)}
	result := mustRunIAMAtomicText(t, ctx, database, "t_iam_a",
		createIAMRBACRuntimeAction, arguments...)
	if result != iamRBACRuntimeActionID+"|1|1|false" {
		t.Fatalf("first RBAC result = %s", result)
	}
	replay := mustRunIAMAtomicText(t, ctx, database, "t_iam_a",
		createIAMRBACRuntimeAction, arguments...)
	if replay != iamRBACRuntimeActionID+"|1|1|true" {
		t.Fatalf("stale-version replay result = %s", replay)
	}
	conflict := append([]any(nil), arguments...)
	conflict[1] = bytes.Repeat([]byte{0x42}, 32)
	if _, err := runIAMAtomicText(ctx, database, "t_iam_a",
		createIAMRBACRuntimeAction, conflict...); err == nil {
		t.Fatal("operation UUID reuse with a different hash was accepted")
	}
	if clock := iamRBACClock(t, ctx, database, "t_iam_a"); clock != 1 {
		t.Fatalf("replay/conflict clock = %d, want 1", clock)
	}
	second := mustRunIAMAtomicText(t, ctx, database, "t_iam_b",
		createIAMRBACRuntimeAction, arguments...)
	if second != iamRBACRuntimeActionID+"|1|1|false" {
		t.Fatalf("second tenant result = %s", second)
	}
	assertIAMRBACRuntimeScopeAndACL(t, ctx, database)
}

func assertIAMRBACRuntimeScopeAndACL(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	_, err := runIAMAtomicText(ctx, database, "t_iam_a", `
SELECT action_key FROM iam_rbac_action_get_v1(
    't_iam_b'::name, $1::uuid,
    '51000000-0000-4000-8000-000000000001',
    '93000000-0000-4000-8000-000000000001')`, iamRBACRuntimeActionID)
	if err == nil {
		t.Fatal("RBAC read accepted a foreign tenant schema")
	}
	connection, transaction, err := beginIAMAtomicRuntimeTx(ctx, database, "t_iam_a")
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	_, err = transaction.ExecContext(ctx, "INSERT INTO iam_actions (id) VALUES ($1)",
		"98200000-0000-4000-8000-000000000001")
	_ = transaction.Rollback()
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "permission") {
		t.Fatalf("runtime direct table DML error = %v", err)
	}
}
