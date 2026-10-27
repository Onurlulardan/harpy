//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"
)

const permissionCacheFlushCall = `SELECT authorization_clock,result_revision,replayed
FROM iam_permission_cache_flush_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    '10000000-0000-4000-8000-000000000001'::uuid,
    $1::uuid,$2::bytea,$3::uuid,$4::bigint,$5::bigint,$6::bigint)`

func TestIAMPermissionCacheFlushAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, origin)
			preparePermissionCacheFlush(t, ctx, database, origin)
			assertPermissionCacheFlushContract(t, ctx, database)
		})
	}
}

func assertPermissionCacheFlushContract(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	operationID := "61024000-0000-4000-8000-000000000001"
	request := bytes.Repeat([]byte{0x61}, 32)
	clock, revision, replayed := callPermissionCacheFlush(t, ctx, database,
		operationID, request, iamUserOneID, 1, 1, 7)
	if clock != 2 || revision != 2 || replayed {
		t.Fatalf("permission cache flush = %d/%d/%t", clock, revision, replayed)
	}
	replayClock, replayRevision, replay := callPermissionCacheFlush(t, ctx, database,
		operationID, request, iamUserOneID, 1, 1, 7)
	if replayClock != clock || replayRevision != revision || !replay {
		t.Fatalf("permission cache replay = %d/%d/%t", replayClock, replayRevision, replay)
	}
	assertPermissionCacheFlushDenied(t, ctx, database, operationID,
		bytes.Repeat([]byte{0x62}, 32), iamUserOneID, 1, 1, 7, "23505")
	assertPermissionCacheFlushDenied(t, ctx, database,
		"61024000-0000-4000-8000-000000000002", request,
		iamUserOneID, 2, 1, 7, "42501")
	assertPermissionCacheFlushDenied(t, ctx, database,
		"61024000-0000-4000-8000-000000000003", request,
		iamUserOneID, 1, 2, 7, "42501")
	assertPermissionCacheFlushDenied(t, ctx, database,
		"61024000-0000-4000-8000-000000000004", request,
		iamUserOneID, 1, 1, 8, "40001")
	assertPermissionCacheFlushDenied(t, ctx, database,
		"61024000-0000-4000-8000-000000000005", request,
		iamUserTwoID, 1, 1, 7, "42501")
	var finalClock, receipts int64
	mustQueryIAMTenantValue(t, ctx, database, "t_iam_a",
		`SELECT clock_value FROM iam_authorization_clock WHERE singleton`, &finalClock)
	mustQueryIAMTenantValue(t, ctx, database, "t_iam_a",
		`SELECT count(*) FROM iam_authorization_operation_receipts`, &receipts)
	if finalClock != 2 || receipts != 1 {
		t.Fatalf("permission cache evidence = clock:%d receipts:%d", finalClock, receipts)
	}
}

func callPermissionCacheFlush(
	t *testing.T, ctx context.Context, database *sql.DB,
	operationID string, request []byte, actor string,
	tokenVersion, authorizationVersion, platformClock int64,
) (int64, int64, bool) {
	t.Helper()
	connection, transaction := beginPermissionCacheRuntimeTx(t, ctx, database)
	defer connection.Close()
	var clock, revision int64
	var replayed bool
	err := transaction.QueryRowContext(ctx, permissionCacheFlushCall,
		operationID, request, actor, tokenVersion, authorizationVersion, platformClock,
	).Scan(&clock, &revision, &replayed)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatalf("permission cache flush: %v", err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return clock, revision, replayed
}
