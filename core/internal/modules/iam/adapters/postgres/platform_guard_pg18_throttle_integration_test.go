//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func assertIAMThrottleGuards(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	refreshSelector []byte,
) {
	t.Helper()
	assertIAMTenantThrottleProgress(t, ctx, database)
	assertIAMEnvironmentThrottle(t, ctx, database)
	assertIAMThrottleEpochCarry(t, ctx, database)
	assertIAMThrottlePolicyRejectsCaller(t, ctx, database)
	assertIAMStaleRefreshTarget(t, ctx, database, refreshSelector)
	assertIAMTenantTarget(t, ctx, database, 2)
}

func assertIAMTenantThrottleProgress(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	hmac := iamBytes(51, 32)
	for attempt := 1; attempt <= iamGuardIdentifierThreshold+1; attempt++ {
		count, blocked, revision, err := recordIAMTenantAttempt(ctx, database.runtime,
			iamGuardTenantA, "progress-key", hmac, 1, iamGuardPolicyVersion,
			iamGuardAuthWindow, iamGuardIdentifierThreshold, iamGuardAuthBlock)
		shouldBlock := attempt >= iamGuardIdentifierThreshold
		if err != nil || count != attempt || revision != int64(attempt) ||
			(blocked != nil) != shouldBlock {
			t.Fatalf("tenant attempt %d=%d/%v/%d err=%v",
				attempt, count, blocked, revision, err)
		}
	}
}

func assertIAMEnvironmentThrottle(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	if _, err := database.migrator.ExecContext(ctx,
		"DELETE FROM platform.environment_identity"); err != nil {
		t.Fatal(err)
	}
	_, _, _, err := recordIAMEnvironmentAttempt(ctx, database.runtime,
		"missing-environment", iamBytes(56, 32), iamGuardPolicyVersion,
		iamGuardAuthWindow, iamGuardIdentifierThreshold, iamGuardAuthBlock)
	assertIAMSQLState(t, err, "P5721")
	if _, err := database.migrator.ExecContext(ctx, `INSERT INTO
platform.environment_identity VALUES(true,'30000000-0000-4000-8000-000000000001')`); err != nil {
		t.Fatal(err)
	}
	setIAMGuardAuthorityState(t, ctx, database, "cutover_groups", "", "v1")
	count, _, _, err := recordIAMEnvironmentAttempt(ctx, database.runtime,
		"fenced-tenant-independent", iamBytes(57, 32), iamGuardPolicyVersion,
		iamGuardAuthWindow, iamGuardIdentifierThreshold, iamGuardAuthBlock)
	setIAMGuardAuthorityState(t, ctx, database, "cutover_groups", "", "v2")
	if err != nil || count != 1 {
		t.Fatalf("environment writer depended on tenant authority: %v", err)
	}
	count, blocked, revision, err := recordIAMEnvironmentAttempt(ctx,
		database.runtime, "environment-key", iamBytes(52, 32), iamGuardPolicyVersion,
		iamGuardAuthWindow, iamGuardIdentifierThreshold, iamGuardAuthBlock)
	if err != nil || count != 1 || blocked != nil || revision != 1 {
		t.Fatalf("environment attempt=%d/%v/%d err=%v", count, blocked, revision, err)
	}
	var scopeID, environmentID string
	if err := database.migrator.QueryRowContext(ctx, `SELECT scope_id,environment_id
FROM platform.iam_auth_throttle_buckets WHERE bucket_scope_kind='environment'`).
		Scan(&scopeID, &environmentID); err != nil || scopeID != environmentID {
		t.Fatalf("derived environment=%s/%s err=%v", scopeID, environmentID, err)
	}
}

func assertIAMThrottleEpochCarry(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	hmac := iamBytes(53, 32)
	var oldCount int
	var oldBlock *time.Time
	var oldRevision int64
	var err error
	for range iamGuardIdentifierThreshold {
		oldCount, oldBlock, oldRevision, err = recordIAMTenantAttempt(ctx,
			database.runtime, iamGuardTenantA, "epoch-key", hmac, 1,
			iamGuardPolicyVersion, iamGuardAuthWindow,
			iamGuardIdentifierThreshold, iamGuardAuthBlock)
		if err != nil {
			t.Fatal(err)
		}
	}
	if oldCount != iamGuardIdentifierThreshold || oldBlock == nil ||
		oldRevision != int64(iamGuardIdentifierThreshold) {
		t.Fatalf("old epoch block=%d/%v/%d err=%v",
			oldCount, oldBlock, oldRevision, err)
	}
	advanceIAMGuardAuthority(t, ctx, database.migrator, iamGuardTenantA, 2)
	count, blocked, revision, err := recordIAMTenantAttempt(ctx, database.runtime,
		iamGuardTenantA, "epoch-key", hmac, 2, iamGuardPolicyVersion,
		iamGuardAuthWindow, iamGuardIdentifierThreshold, iamGuardAuthBlock)
	if err != nil || count != oldCount+1 || blocked == nil ||
		blocked.Before(*oldBlock) || revision != 1 {
		t.Fatalf("new epoch carry=%d/%v/%d err=%v", count, blocked, revision, err)
	}
	var rows int
	var retainedRevision int64
	if err := database.migrator.QueryRowContext(ctx, `SELECT count(*),
max(revision) FILTER (WHERE authority_epoch=1)
FROM platform.iam_auth_throttle_buckets
WHERE bucket_scope_kind='tenant' AND scope_id=$1 AND key_id='epoch-key'`,
		iamGuardTenantA).Scan(&rows, &retainedRevision); err != nil ||
		rows != 2 || retainedRevision != oldRevision {
		t.Fatalf("retained epochs=%d revision=%d err=%v", rows, retainedRevision, err)
	}
	_, _, _, err = recordIAMTenantAttempt(ctx, database.runtime,
		iamGuardTenantA, "stale-key", iamBytes(54, 32), 1,
		iamGuardPolicyVersion, iamGuardAuthWindow,
		iamGuardIdentifierThreshold, iamGuardAuthBlock)
	assertIAMSQLState(t, err, "P5701")
}

func assertIAMThrottlePolicyRejectsCaller(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	_, _, _, err := recordIAMTenantAttempt(ctx, database.runtime,
		iamGuardTenantA, "policy-key", iamBytes(55, 32), 2,
		iamGuardPolicyVersion, iamGuardAuthWindow, 1, iamGuardAuthBlock)
	assertIAMSQLState(t, err, "42501")
	err = database.runtime.QueryRowContext(ctx, iamTenantThrottleSQL,
		iamGuardTenantA, "login", "tenant_slug", "pair-key", iamBytes(56, 32),
		int64(2), int64(2), iamGuardPolicyVersion, iamGuardAuthWindow, 30,
		iamGuardAuthBlock).Scan(new(int), new(*time.Time), new(int64))
	assertIAMSQLState(t, err, "42501")
}

func assertIAMStaleRefreshTarget(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	selector []byte,
) {
	t.Helper()
	err := database.runtime.QueryRowContext(ctx, `SELECT * FROM
platform.iam_resolve_refresh_target_v2($1)`, selector).
		Scan(new(string), new(string), new(string), new([]byte), new(int64),
			new([]byte), new(int64), new(int64), new(int64))
	assertIAMSQLState(t, err, "P5732")
}
