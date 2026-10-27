//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

const authHTTPThrottleKID = "ABCDEFGHIJKLMNOP"

const authHTTPThrottleRecordSQL = `SELECT platform.iam_auth_throttle_record_failure_v2(
$1,$2,$3,$4,$5,$6,$7,$8)`

const authHTTPThrottleBlockedSQL = `SELECT platform.iam_auth_throttle_blocked_v2(
$1,$2,$3,$4,$5,$6,$7,$8)`

func TestIAMAuthHTTPThrottlePostgreSQL18(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx := context.Background()
			database := startIAMPlatformGuardPG18(t, ctx, origin)
			prepareIAMPlatformGuardSchema(t, ctx, database, origin)
			applyAuthHTTPPlatformMigrations(t, ctx, database, origin)
			assertAuthHTTPEnvironmentThrottle(t, ctx, database)
			assertAuthHTTPTenantThrottle(t, ctx, database)
			assertAuthHTTPThrottleACL(t, ctx, database)
			assertAuthHTTPLoginTargetKeyContract(t, ctx, database)
		})
	}
}

func applyAuthHTTPPlatformMigrations(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase, origin string,
) {
	t.Helper()
	connection, err := database.migrator.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	for _, name := range []string{
		"20260716014602_create_iam_hmac_key_id_valid_v1",
		"20260716014603_guard_iam_auth_hmac_key_ids_v1",
		"20260716014604_create_iam_auth_throttle_blocked_v2",
		"20260716014605_create_iam_auth_throttle_record_failure_v2",
		"20260716014606_grant_iam_auth_throttle_runtime_v2",
		"20260716014607_guard_iam_directory_hmac_key_ids_v1",
		"20260725124300_remove_auth_throttle_read_lock",
		loginTargetKeyContractMigration,
		"20260725124610_create_iam_login_target_resolver_v3",
		"20260725124710_create_iam_login_target_resolver_v4",
		"20260725124810_create_iam_login_target_resolver_v5",
		"20260725124910_create_iam_login_target_resolver_v6",
		"20260725125000_remove_iam_tenant_authority_read_lock",
		"20260725125110_create_iam_login_target_resolver_v7",
	} {
		applyIAMPlatformGuardDocument(
			t, ctx, connection, name, readIAMPlatformAuthMigration(t, origin, name),
		)
	}
}

func assertAuthHTTPLoginTargetKeyContract(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	hmac := iamBytes(94, 32)
	_, err := database.migrator.ExecContext(ctx, `
INSERT INTO platform.iam_login_directory(
 tenant_id,identifier_kind,lookup_key_id,identifier_hmac,user_id,
 source_identifier_id,source_revision,state,source_sha256,projected_at,tombstoned_at)
VALUES($1,'email_v1',$2,$3,$4,$5,1,'active',$6,$7,NULL)`,
		iamGuardTenantA, authHTTPThrottleKID, hmac, uuid.New(), uuid.New(),
		iamBytes(95, 32), time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	err = database.runtime.QueryRowContext(ctx, `SELECT tenant_id FROM
platform.iam_resolve_login_target_v7($1,'email_v1',$2,$3)`,
		iamGuardTenantA, authHTTPThrottleKID, hmac).Scan(new(string))
	if err != nil {
		t.Fatalf("canonical Base64URL login key was rejected: %v", err)
	}
	err = database.runtime.QueryRowContext(ctx, `SELECT tenant_id FROM
platform.iam_resolve_login_target_v7($1,'email_v1','lowercase-key',$2)`,
		iamGuardTenantA, hmac).Scan(new(string))
	assertIAMSQLState(t, err, "42501")
}

func assertAuthHTTPEnvironmentThrottle(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	hmac := iamBytes(91, 32)
	runAuthHTTPThrottleRace(t, func() error {
		return database.runtime.QueryRowContext(ctx, authHTTPThrottleRecordSQL,
			"environment", nil, "login", "identifier", authHTTPThrottleKID,
			hmac, nil, nil).Scan(new(bool))
	})
	var count int
	var blocked bool
	err := database.migrator.QueryRowContext(ctx, `SELECT attempt_count,
blocked_until > clock_timestamp() FROM platform.iam_auth_throttle_buckets
WHERE bucket_scope_kind='environment' AND purpose='login'
AND dimension='identifier' AND key_id=$1 AND bucket_hmac=$2`,
		authHTTPThrottleKID, hmac).Scan(&count, &blocked)
	if err != nil || count != 5 || !blocked {
		t.Fatalf("environment throttle = %d/%t: %v", count, blocked, err)
	}
}

func assertAuthHTTPTenantThrottle(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	hmac := iamBytes(92, 32)
	runAuthHTTPThrottleRace(t, func() error {
		return database.runtime.QueryRowContext(ctx, authHTTPThrottleRecordSQL,
			"tenant", iamGuardTenantA, "login", "identifier", authHTTPThrottleKID,
			hmac, int64(1), int64(1)).Scan(new(bool))
	})
	var count int
	var blocked bool
	err := database.migrator.QueryRowContext(ctx, `SELECT attempt_count,
blocked_until > clock_timestamp() FROM platform.iam_auth_throttle_buckets
WHERE tenant_id=$1 AND purpose='login' AND dimension='identifier'
AND key_id=$2 AND bucket_hmac=$3`, iamGuardTenantA,
		authHTTPThrottleKID, hmac).Scan(&count, &blocked)
	if err != nil || count != 5 || !blocked {
		t.Fatalf("tenant A throttle = %d/%t: %v", count, blocked, err)
	}
	var tenantBBlocked bool
	err = database.runtime.QueryRowContext(ctx, authHTTPThrottleRecordSQL,
		"tenant", iamGuardTenantB, "login", "identifier", authHTTPThrottleKID,
		hmac, int64(1), int64(1)).Scan(&tenantBBlocked)
	if err != nil || tenantBBlocked {
		t.Fatalf("tenant B shared tenant A bucket: %t/%v", tenantBBlocked, err)
	}
}

func assertAuthHTTPThrottleACL(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	hmac := iamBytes(93, 32)
	var blocked bool
	err := database.runtime.QueryRowContext(ctx, authHTTPThrottleBlockedSQL,
		"environment", nil, "refresh", "ip", authHTTPThrottleKID,
		hmac, nil, nil).Scan(&blocked)
	if err != nil || blocked {
		t.Fatalf("runtime throttle probe = %t/%v", blocked, err)
	}
	transaction, err := database.runtime.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	err = transaction.QueryRowContext(ctx, authHTTPThrottleBlockedSQL, "environment",
		nil, "refresh", "ip", authHTTPThrottleKID, hmac, nil, nil).Scan(&blocked)
	if err != nil || blocked {
		t.Fatalf("read-only runtime throttle probe = %t/%v", blocked, err)
	}
	err = transaction.QueryRowContext(ctx, authHTTPThrottleBlockedSQL, "tenant",
		iamGuardTenantA, "login", "identifier", authHTTPThrottleKID, hmac,
		int64(1), int64(1)).Scan(&blocked)
	if err != nil {
		t.Fatalf("read-only tenant throttle probe failed: %v", err)
	}
	err = database.adopter.QueryRowContext(ctx, authHTTPThrottleBlockedSQL, "environment",
		nil, "refresh", "ip", authHTTPThrottleKID, hmac, nil, nil).Scan(&blocked)
	assertIAMSQLState(t, err, "42501")
	err = database.runtime.QueryRowContext(ctx, authHTTPThrottleBlockedSQL, "environment",
		nil, "refresh", "ip", "lowercase-key", hmac, nil, nil).Scan(&blocked)
	assertIAMSQLState(t, err, "42501")
}
