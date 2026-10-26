//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
)

func assertIAMInitialPasswordAttemptLimit(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	fixture iamInitialPasswordFixture,
) {
	t.Helper()
	revision := claimIAMInitialPassword(t, ctx, database, fixture)
	for attempt := int16(1); attempt <= 5; attempt++ {
		transaction := beginInitialPasswordRuntimeTx(t, ctx, database, fixture)
		var state string
		var attempts int16
		var nextRevision int64
		err := transaction.QueryRowContext(ctx, `SELECT * FROM iam_initial_password_reject_v1(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5)`,
			fixture.tenant.id, fixture.selector, fixture.secretHMAC,
			fixture.csrfHMAC, revision,
		).Scan(&state, &attempts, &nextRevision)
		if err != nil || attempts != attempt || nextRevision != revision+1 ||
			(attempt < 5 && state != "active") || (attempt == 5 && state != "exhausted") {
			_ = transaction.Rollback()
			t.Fatalf("attempt %d state=%s/%d/%d err=%v",
				attempt, state, attempts, nextRevision, err)
		}
		if err := transaction.Commit(); err != nil {
			t.Fatal(err)
		}
		revision = nextRevision
	}
	transaction := beginInitialPasswordRuntimeTx(t, ctx, database, fixture)
	err := transaction.QueryRowContext(ctx, `SELECT * FROM iam_initial_password_reject_v1(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5)`,
		fixture.tenant.id, fixture.selector, fixture.secretHMAC,
		fixture.csrfHMAC, revision,
	).Scan(new(string), new(int16), new(int64))
	_ = transaction.Rollback()
	assertIAMSQLState(t, err, "40001")
}

func assertIAMInitialPasswordFinalize(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	fixture iamInitialPasswordFixture,
) {
	t.Helper()
	revision := claimIAMInitialPassword(t, ctx, database, fixture)
	transaction := beginInitialPasswordRuntimeTx(t, ctx, database, fixture)
	finalizeIAMInitialPassword(t, ctx, transaction, fixture, revision)
	if err := transaction.Rollback(); err != nil {
		t.Fatal(err)
	}
	assertIAMInitialCredentialState(t, ctx, database, fixture, "active", "reset_required", 1)
	transaction = beginInitialPasswordRuntimeTx(t, ctx, database, fixture)
	finalizeIAMInitialPassword(t, ctx, transaction, fixture, revision)
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	assertIAMInitialCredentialState(t, ctx, database, fixture, "retired", "active", 2)
	transaction = beginInitialPasswordRuntimeTx(t, ctx, database, fixture)
	err := finalizeIAMInitialPasswordError(ctx, transaction, fixture, revision)
	_ = transaction.Rollback()
	assertIAMSQLState(t, err, "40001")
}

func beginInitialPasswordRuntimeTx(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	fixture iamInitialPasswordFixture,
) *sql.Tx {
	t.Helper()
	transaction, err := beginIAMAuthRuntimeTx(ctx, database.runtime, fixture.tenant.schema)
	if err != nil {
		t.Fatal(err)
	}
	return transaction
}

func finalizeIAMInitialPassword(
	t *testing.T, ctx context.Context, transaction *sql.Tx,
	fixture iamInitialPasswordFixture, revision int64,
) {
	t.Helper()
	if err := finalizeIAMInitialPasswordError(ctx, transaction, fixture, revision); err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
}

func finalizeIAMInitialPasswordError(
	ctx context.Context, transaction *sql.Tx,
	fixture iamInitialPasswordFixture, revision int64,
) error {
	return transaction.QueryRowContext(ctx, `SELECT * FROM iam_initial_password_complete_v1(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5,$6,$7,1,1,$8,$9,$10)`,
		fixture.tenant.id, fixture.selector, fixture.secretHMAC, fixture.csrfHMAC,
		revision, iamUserOneID, fixture.credentialID,
		bytes.Repeat([]byte{0x81}, 16), bytes.Repeat([]byte{0x91}, 32),
		bytes.Repeat([]byte{0xa1}, 32),
	).Scan(new(string), new(string), new(int64), new(int64), new(int64), new(int64),
		new(int64), new(int64), new(int64), new(int64))
}

func assertIAMInitialCredentialState(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	fixture iamInitialPasswordFixture, markerState, credentialState string,
	credentialVersion int64,
) {
	t.Helper()
	var marker, credential string
	var version int64
	connection, transaction := beginIAMTenantTx(
		t, ctx, database.migrator, fixture.tenant.schema,
	)
	defer connection.Close()
	err := transaction.QueryRowContext(ctx, `
SELECT marker.state,credential.state,credential.credential_version
FROM iam_installation_password_credentials AS marker
JOIN iam_password_credentials AS credential ON credential.user_id=marker.user_id
WHERE marker.credential_id=$1`, fixture.credentialID).Scan(&marker, &credential, &version)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	if marker != markerState || credential != credentialState || version != credentialVersion {
		t.Fatalf("credential state=%s/%s/%d", marker, credential, version)
	}
}
