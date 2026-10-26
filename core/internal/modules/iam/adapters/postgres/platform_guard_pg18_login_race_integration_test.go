//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

const iamLoginAdvisoryLockSQL = `SELECT pg_catalog.pg_advisory_xact_lock(
pg_catalog.hashtextextended('harpy:iam-login-target:' || $1::text || ':' ||
$2 || ':' || $3 || ':' || pg_catalog.encode($4,'hex'),0))`

func assertIAMLoginResolverRace(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	hmac := iamBytes(41, 32)
	writer, err := database.migrator.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.ExecContext(ctx, iamLoginAdvisoryLockSQL,
		iamGuardTenantB, "email_v1", "guard-key", hmac); err != nil {
		t.Fatal(err)
	}
	_, err = writer.ExecContext(ctx, `INSERT INTO platform.iam_login_directory(
tenant_id,identifier_kind,lookup_key_id,identifier_hmac,user_id,
source_identifier_id,source_revision,state,source_sha256,projected_at,tombstoned_at)
VALUES($1,'email_v1','guard-key',$2,$3,$4,6,'active',$5,$6,NULL)`,
		iamGuardTenantB, hmac, uuid.New(), uuid.New(), iamBytes(6, 32), time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	assertIAMLoginTarget(t, ctx, database, iamGuardTenantA, hmac, 4, 2)
	if err := writer.Commit(); err != nil {
		t.Fatal(err)
	}
	assertIAMLoginTarget(t, ctx, database, iamGuardTenantB, hmac, 6, 3)
	assertIAMSameTenantLoginSerialization(t, ctx, database, hmac)
	if _, err := database.migrator.ExecContext(ctx, `DELETE FROM
platform.iam_login_directory WHERE tenant_id=$1 AND lookup_key_id='guard-key'`,
		iamGuardTenantB); err != nil {
		t.Fatal(err)
	}
	assertIAMLoginValidationPrecedesLock(t, ctx, database)
}

func assertIAMSameTenantLoginSerialization(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase, hmac []byte,
) {
	t.Helper()
	writer, err := database.migrator.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.ExecContext(ctx, iamLoginAdvisoryLockSQL,
		iamGuardTenantA, "email_v1", "guard-key", hmac); err != nil {
		t.Fatal(err)
	}
	_, err = writer.ExecContext(ctx, `UPDATE platform.iam_login_directory
SET source_revision=7,source_sha256=$2,projected_at=$3
WHERE tenant_id=$1 AND lookup_key_id='guard-key'`,
		iamGuardTenantA, iamBytes(7, 32), time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	var tenant string
	var userID, identifierID, identifierKind string
	var revision, group, iam, layout int64
	var sourceHash []byte
	go func() {
		done <- database.runtime.QueryRowContext(ctx, `SELECT * FROM
platform.iam_resolve_login_target_v2($1,'email_v1','guard-key',$2)`,
			iamGuardTenantA, hmac).Scan(
			&tenant, &userID, &identifierID, &identifierKind,
			&revision, &sourceHash, &group, &iam, &layout)
	}()
	assertIAMOperationBlocks(t, done)
	if err := writer.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := awaitIAMGuardResult(t, done); err != nil || tenant != iamGuardTenantA ||
		revision != 7 || identifierKind != "email_v1" ||
		group != 2 || iam != 2 || layout != 2 {
		t.Fatalf("same-tenant login serialization=%s/%d/%d/%d/%d err=%v",
			tenant, revision, group, iam, layout, err)
	}
}

func assertIAMLoginValidationPrecedesLock(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	transaction, err := database.runtime.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	invalidHMAC := iamBytes(83, 31)
	if _, err := transaction.ExecContext(ctx, iamLoginAdvisoryLockSQL,
		iamGuardTenantA, "email_v1", "invalid-key", invalidHMAC); err != nil {
		t.Fatal(err)
	}
	err = database.runtime.QueryRowContext(ctx, `SELECT * FROM
platform.iam_resolve_login_target_v2($1,'email_v1','invalid-key',$2)`,
		iamGuardTenantA, invalidHMAC).
		Scan(new(string), new(string), new(string), new(string), new(int64),
			new([]byte), new(int64), new(int64), new(int64))
	assertIAMSQLState(t, err, "42501")
}
