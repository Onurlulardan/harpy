//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"
)

func assertPurgeWithoutTrustedProofRejected(t *testing.T, ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	_, err := transaction.ExecContext(ctx, `SELECT *
FROM platform.tenancy_customer_lifecycle_accept_v1(
$1,$2,$3,$4,'purge',2,'tenant-purge-no-proof-01',$5,
NULL,NULL,NULL,NULL,NULL,$6)`, testEnvironmentID, lifecycleControlID,
		lifecycleActorID, lifecyclePurgeID, bytes.Repeat([]byte{0x71}, 32),
		time.Now().UTC().Truncate(time.Microsecond))
	_ = transaction.Rollback()
	if err == nil {
		t.Fatal("purge without trusted proof was accepted")
	}
}

func assertLifecycleTenantTimes(t *testing.T, ctx context.Context, database *sql.DB,
	tenantID string, wantOffboarded, wantPurged bool, wantRevision int64,
) {
	t.Helper()
	var offboarded, purged sql.NullTime
	var deleted sql.NullTime
	var revision int64
	var status string
	err := database.QueryRowContext(ctx, `SELECT offboarded_at,purged_at,deleted_at,
lifecycle_revision,status FROM platform.tenants WHERE id=$1`, tenantID).Scan(
		&offboarded, &purged, &deleted, &revision, &status)
	if err != nil || offboarded.Valid != wantOffboarded || purged.Valid != wantPurged ||
		deleted.Valid != wantOffboarded || revision != wantRevision || status != "SUSPENDED" {
		t.Fatalf("tenant lifecycle offboard=%t purge=%t deleted=%t revision=%d status=%s err=%v",
			offboarded.Valid, purged.Valid, deleted.Valid, revision, status, err)
	}
}
