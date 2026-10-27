//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

const iamThrottleFutureProbeSQL = `
CREATE FUNCTION platform.test_iam_throttle_future_v1(p_key text)
RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog AS $f$
BEGIN
 IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
    CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
  RAISE EXCEPTION USING ERRCODE='42501';
 END IF;
 UPDATE platform.iam_auth_throttle_buckets AS bucket
 SET attempt_count=bucket.attempt_count+1,
     last_attempt_at=bucket.last_attempt_at+interval '2 seconds',
     revision=bucket.revision+1
 WHERE bucket.bucket_scope_kind='tenant' AND bucket.key_id=p_key;
END;$f$;
REVOKE ALL ON FUNCTION platform.test_iam_throttle_future_v1(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.test_iam_throttle_future_v1(text)
TO harpy_runtime;`

func assertIAMThrottleClockRollback(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	hmac := iamBytes(91, 32)
	_, _, _, err := recordIAMTenantAttempt(ctx, database.runtime,
		iamGuardTenantA, "clock-key", hmac, 2, iamGuardPolicyVersion,
		iamGuardAuthWindow, iamGuardIdentifierThreshold, iamGuardAuthBlock)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.migrator.ExecContext(ctx, iamThrottleFutureProbeSQL); err != nil {
		t.Fatal(err)
	}
	if _, err := database.runtime.ExecContext(ctx,
		"SELECT platform.test_iam_throttle_future_v1('clock-key')"); err != nil {
		t.Fatal(err)
	}
	var priorAttempt int
	var priorRevision int64
	var priorLast time.Time
	if err := database.migrator.QueryRowContext(ctx, `SELECT attempt_count,
last_attempt_at,revision FROM platform.iam_auth_throttle_buckets
WHERE key_id='clock-key' AND authority_epoch=2`).
		Scan(&priorAttempt, &priorLast, &priorRevision); err != nil {
		t.Fatal(err)
	}
	advanceIAMGuardAuthority(t, ctx, database.migrator, iamGuardTenantA, 3)
	count, _, revision, err := recordIAMTenantAttempt(ctx, database.runtime,
		iamGuardTenantA, "clock-key", hmac, 3, iamGuardPolicyVersion,
		iamGuardAuthWindow, iamGuardIdentifierThreshold, iamGuardAuthBlock)
	if err != nil || count != priorAttempt+1 || revision != 1 || priorRevision != 2 {
		t.Fatalf("clock rollback carry=%d/%d prior=%d/%d err=%v",
			count, revision, priorAttempt, priorRevision, err)
	}
	var currentLast time.Time
	if err := database.migrator.QueryRowContext(ctx, `SELECT last_attempt_at
FROM platform.iam_auth_throttle_buckets
WHERE key_id='clock-key' AND authority_epoch=3`).Scan(&currentLast); err != nil {
		t.Fatal(err)
	}
	if !currentLast.After(priorLast) {
		t.Fatalf("new epoch last_attempt_at=%v, prior=%v", currentLast, priorLast)
	}
	if _, err := database.migrator.ExecContext(ctx,
		"DROP FUNCTION platform.test_iam_throttle_future_v1(text)"); err != nil {
		t.Fatal(err)
	}
}
