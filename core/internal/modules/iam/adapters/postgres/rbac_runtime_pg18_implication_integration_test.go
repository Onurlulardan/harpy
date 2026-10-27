//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
)

const iamRBACRuntimeImplicationID = "99000000-0000-4000-8000-000000000001"

func assertIAMRBACImplicationLifecycle(t *testing.T, ctx context.Context,
	database *sql.DB, schema string) {
	t.Helper()
	create := mustRunIAMAtomicText(t, ctx, database, schema, `
SELECT pg_catalog.concat_ws('|', result_id::text, result_revision::text,
    authorization_clock::text, replayed::text)
FROM iam_rbac_implication_create_v1(
    pg_catalog.current_schema()::name, $1::uuid, $2::bytea, $3::uuid,
    '93000000-0000-4000-8000-000000000001',
    '93000000-0000-4000-8000-000000000002', 'forms.parent.v1',
    '51000000-0000-4000-8000-000000000001',
    '93000000-0000-4000-8000-000000000001', 1, 2)`,
		"99100000-0000-4000-8000-000000000001", bytes.Repeat([]byte{0x61}, 32),
		iamRBACRuntimeImplicationID)
	if create != iamRBACRuntimeImplicationID+"|1|2|false" {
		t.Fatalf("implication create = %s", create)
	}
	assertIAMPermissionDecision(t, ctx, database, schema, true)
	hit := mustRunIAMAtomicText(t, ctx, database, schema, `
SELECT pg_catalog.concat_ws('|', result_id::text, result_revision::text,
    authorization_clock::text, replayed::text)
FROM iam_rbac_implication_record_hit_v1(
    pg_catalog.current_schema()::name, $1::uuid, $2::bytea, $3::uuid, 1,
    '51000000-0000-4000-8000-000000000001',
    '93000000-0000-4000-8000-000000000001', 1, 3)`,
		"99100000-0000-4000-8000-000000000002", bytes.Repeat([]byte{0x62}, 32),
		iamRBACRuntimeImplicationID)
	if hit != iamRBACRuntimeImplicationID+"|2|3|false" {
		t.Fatalf("implication hit = %s", hit)
	}
	retireArgs := []any{"99100000-0000-4000-8000-000000000003",
		bytes.Repeat([]byte{0x63}, 32), iamRBACRuntimeImplicationID,
		"99200000-0000-4000-8000-000000000001"}
	retire := mustRunIAMAtomicText(t, ctx, database, schema, `
SELECT pg_catalog.concat_ws('|', result_id::text, result_revision::text,
    authorization_clock::text, replayed::text)
FROM iam_rbac_implication_retire_v1(
    pg_catalog.current_schema()::name, $1::uuid, $2::bytea, $3::uuid, 2, $4::uuid,
    '51000000-0000-4000-8000-000000000001',
    '93000000-0000-4000-8000-000000000001', 1, 3)`, retireArgs...)
	if retire != iamRBACRuntimeImplicationID+"|3|4|false" {
		t.Fatalf("implication retire = %s", retire)
	}
	assertIAMPermissionDecision(t, ctx, database, schema, false)
	replay := mustRunIAMAtomicText(t, ctx, database, schema, `
SELECT replayed::text FROM iam_rbac_implication_retire_v1(
    pg_catalog.current_schema()::name, $1::uuid, $2::bytea, $3::uuid, 2, $4::uuid,
    '51000000-0000-4000-8000-000000000001',
    '93000000-0000-4000-8000-000000000001', 1, 3)`, retireArgs...)
	if replay != "true" || iamRBACClock(t, ctx, database, schema) != 4 {
		t.Fatal("retired implication replay changed authorization state")
	}
}

func assertIAMPermissionDecision(t *testing.T, ctx context.Context, database *sql.DB,
	schema string, want bool) {
	t.Helper()
	var allowed bool
	mustQueryIAMTenantValue(t, ctx, database, schema, `SELECT iam_user_has_permission_v1(
    '51000000-0000-4000-8000-000000000001',
    '93000000-0000-4000-8000-000000000002')`, &allowed)
	if allowed != want {
		t.Fatalf("implication permission decision = %t, want %t", allowed, want)
	}
}
