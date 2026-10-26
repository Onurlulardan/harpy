//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
)

const revokeIAMLifecycleSession = `
SELECT pg_catalog.concat_ws('|', result_id::text, result_revision::text,
    replayed::text)
FROM iam_session_revoke_one_v1(
    pg_catalog.current_schema()::name, $1::uuid, $2::uuid, $3::bytea,
    $4::uuid, $5::bigint, $6::uuid, $7::uuid, $8::bigint, $9::bigint
)`

func assertIAMLifecycleRuntimeBehavior(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		assertIAMLifecycleReplay(t, ctx, database, tenant)
		assertIAMLifecycleStaleCAS(t, ctx, database, tenant)
		assertIAMLifecycleSameOperationRace(t, ctx, database, tenant)
		assertIAMLifecycleRollback(t, ctx, database, tenant)
	}
}

func assertIAMLifecycleReplay(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant iamFoundationTenant,
) {
	t.Helper()
	actor := readIAMIdentityActor(t, ctx, database, tenant.schema, iamUserOneID)
	hash := bytes.Repeat([]byte{0x72}, 32)
	arguments := lifecycleRevokeArguments(tenant, iamLifecycleReplayOperation,
		hash, iamLifecycleReplaySession, 1, actor)
	want := iamLifecycleReplaySession + "|2|false"
	if got := mustRunIAMAtomicText(t, ctx, database, tenant.schema,
		revokeIAMLifecycleSession, arguments...); got != want {
		t.Fatalf("%s lifecycle first result=%q want=%q", tenant.schema, got, want)
	}
	want = iamLifecycleReplaySession + "|2|true"
	if got := mustRunIAMAtomicText(t, ctx, database, tenant.schema,
		revokeIAMLifecycleSession, arguments...); got != want {
		t.Fatalf("%s lifecycle replay=%q want=%q", tenant.schema, got, want)
	}
	arguments[2] = bytes.Repeat([]byte{0x73}, 32)
	_, err := runIAMAtomicText(ctx, database, tenant.schema,
		revokeIAMLifecycleSession, arguments...)
	assertIAMIdentitySQLState(t, err, "23505")
	assertIAMLifecycleState(t, ctx, database, tenant.schema,
		iamLifecycleReplaySession, "revoked", 2)
}

func assertIAMLifecycleStaleCAS(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant iamFoundationTenant,
) {
	t.Helper()
	actor := readIAMIdentityActor(t, ctx, database, tenant.schema, iamUserOneID)
	arguments := lifecycleRevokeArguments(tenant, iamLifecycleStaleOperation,
		bytes.Repeat([]byte{0x74}, 32), iamLifecycleStaleSession, 1, actor)
	_, err := runIAMAtomicText(ctx, database, tenant.schema,
		revokeIAMLifecycleSession, arguments...)
	assertIAMIdentitySQLState(t, err, "40001")
	assertIAMLifecycleState(t, ctx, database, tenant.schema,
		iamLifecycleStaleSession, "active", 2)
	assertIAMLifecycleReceiptCount(t, ctx, database, tenant.schema,
		iamLifecycleStaleOperation, 0)
}

func lifecycleRevokeArguments(
	tenant iamFoundationTenant, operation string, hash []byte,
	session string, revision int64, actor iamIdentityActor,
) []any {
	return []any{tenant.id, operation, hash, session, revision, iamUserOneID,
		iamIdentityPermissionID, actor.tokenVersion, actor.authorizationVersion}
}
