//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

const (
	iamGuardTenantA             = "10000000-0000-4000-8000-000000000001"
	iamGuardTenantB             = "20000000-0000-4000-8000-000000000002"
	iamGuardPolicyVersion       = 1
	iamGuardAuthWindow          = 300
	iamGuardIdentifierThreshold = 5
	iamGuardAuthBlock           = 900
)

const iamCreateLocatorSQL = `SELECT platform.iam_refresh_locator_create_v1(
$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`

const iamAdvanceLocatorSQL = `SELECT platform.iam_refresh_locator_advance_v1(
$1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`

const iamTenantThrottleSQL = `SELECT result_attempt_count,result_blocked_until,
result_revision FROM platform.iam_tenant_throttle_record_attempt_v1(
$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`

const iamEnvironmentThrottleSQL = `SELECT result_attempt_count,result_blocked_until,
result_revision FROM platform.iam_environment_throttle_record_attempt_v1(
$1,$2,$3,$4,$5,$6,$7,$8)`

type iamPlatformExecer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

type iamPlatformQueryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func iamBytes(value byte, length int) []byte {
	return bytes.Repeat([]byte{value}, length)
}

func createIAMGuardLocator(
	ctx context.Context, database iamPlatformExecer, selector []byte,
	tenantID string, generation int64, sourceState string,
	issued, projected, familyExpiry, retainUntil time.Time,
) error {
	_, err := database.ExecContext(ctx, iamCreateLocatorSQL,
		selector, tenantID, uuid.New(), uuid.New(), generation, generation,
		int64(1), iamBytes(selector[0], 32), sourceState,
		issued, projected, familyExpiry, retainUntil)
	return err
}

func advanceIAMGuardLocator(
	ctx context.Context, database iamPlatformQueryRower, selector []byte,
	tenantID string, generation, expected, next int64,
	sourceState string, projected, retainUntil time.Time,
) (int64, error) {
	var revision int64
	err := database.QueryRowContext(ctx, iamAdvanceLocatorSQL,
		selector, tenantID, generation, generation, expected, next,
		iamBytes(byte(next), 32), sourceState, projected, retainUntil).Scan(&revision)
	return revision, err
}

func recordIAMTenantAttempt(
	ctx context.Context, database iamPlatformQueryRower, tenantID, keyID string,
	hmac []byte, generation int64, policy, window, threshold, block int,
) (int, *time.Time, int64, error) {
	var count int
	var blocked sql.NullTime
	var revision int64
	err := database.QueryRowContext(ctx, iamTenantThrottleSQL,
		tenantID, "login", "identifier", keyID, hmac, generation, generation,
		policy, window, threshold, block).Scan(&count, &blocked, &revision)
	if !blocked.Valid {
		return count, nil, revision, err
	}
	return count, &blocked.Time, revision, err
}

func recordIAMEnvironmentAttempt(
	ctx context.Context, database iamPlatformQueryRower, keyID string,
	hmac []byte, policy, window, threshold, block int,
) (int, *time.Time, int64, error) {
	var count int
	var blocked sql.NullTime
	var revision int64
	err := database.QueryRowContext(ctx, iamEnvironmentThrottleSQL,
		"login", "identifier", keyID, hmac, policy, window, threshold, block).
		Scan(&count, &blocked, &revision)
	if !blocked.Valid {
		return count, nil, revision, err
	}
	return count, &blocked.Time, revision, err
}

func advanceIAMGuardAuthority(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenantID string, generation int64,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err := transaction.ExecContext(ctx, `
UPDATE platform.cutover_groups SET generation=$2
WHERE scope_kind='tenant' AND scope_id=$1 AND group_id='iam_system_resources'`,
		tenantID, generation); err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.ExecContext(ctx, `
UPDATE platform.domain_cutovers SET generation=$2
WHERE scope_kind='tenant' AND scope_id=$1
AND domain IN ('iam','system_resource_layout')`, tenantID, generation); err != nil {
		t.Fatal(err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}
