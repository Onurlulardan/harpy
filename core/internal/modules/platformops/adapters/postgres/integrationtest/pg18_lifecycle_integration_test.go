//go:build integration

package postgres_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

const (
	environmentID = "11111111-1111-4111-8111-111111111111"
	primaryID     = "22222222-2222-4222-8222-222222222222"
	ownerID       = "33333333-3333-4333-8333-333333333333"
	lineageID     = "44444444-4444-4444-8444-444444444444"
	hashA         = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	hashB         = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func exerciseRecoveryLifecycle(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	recordSigningBundle(t, ctx, database, now)
	source := "55555555-5555-4555-8555-555555555555"
	safety := "66666666-6666-4666-8666-666666666666"
	protectRecoverySet(t, ctx, database, source, "77777777-7777-4777-8777-777777777777", now)
	protectRecoverySet(t, ctx, database, safety, "88888888-8888-4888-8888-888888888888", now.Add(time.Minute))
	assertRecoveryPrivileges(t, ctx, database)
	requestReadableRestore(t, ctx, database, source, safety, now.Add(2*time.Minute))
}

func recordSigningBundle(t *testing.T, ctx context.Context, database *sql.DB, now time.Time) {
	t.Helper()
	_, err := database.ExecContext(ctx, `SELECT platform.platformops_record_signing_bundle_v1(
		$1, 1, $2, $3, $4)`, environmentID, hashA, []byte(`{"generation":1}`), now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `SELECT platform.platformops_record_signing_bundle_v1(
		$1, 3, $2, $3, $4)`, environmentID, hashB, []byte(`bad`), now); err == nil {
		t.Fatal("non-monotonic signing generation was accepted")
	}
}

func protectRecoverySet(
	t *testing.T, ctx context.Context, database *sql.DB,
	recoverySetID, operationID string, now time.Time,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `SELECT platform.platformops_request_recovery_set_v2(
		$1,$2,$3,$4)`, recoverySetID, operationID, environmentID, now)
	if err != nil {
		t.Fatal(err)
	}
	for _, transition := range [][2]string{
		{"requested", "holding"}, {"holding", "snapshotting"},
		{"snapshotting", "copying"}, {"copying", "verifying"},
	} {
		assertBooleanRoutine(t, ctx, database,
			`SELECT platform.platformops_advance_recovery_set_v1($1,$2,$3,NULL,$4)`,
			recoverySetID, transition[0], transition[1], now)
	}
	kinds := []string{"database_archive", "documents_inventory", "key_custody", "owner_anchor", "protected_authority"}
	for index, kind := range kinds {
		_, err := database.ExecContext(ctx, `SELECT platform.platformops_record_recovery_component_v1(
			$1,$2,$3,$4,$5,$6,$7)`, recoverySetID, kind,
			fmt.Sprintf("objects/%s/%s", recoverySetID, kind), fmt.Sprintf("v%d", index+1),
			hashB, int64(index+1), now)
		if err != nil {
			t.Fatal(err)
		}
	}
	var databaseSystemID string
	if err := database.QueryRowContext(ctx, `SELECT database_system_id
		FROM platform.platformops_current_environment_identity_v1($1)`, environmentID).
		Scan(&databaseSystemID); err != nil {
		t.Fatal(err)
	}
	assertBooleanRoutine(t, ctx, database,
		`SELECT platform.platformops_complete_recovery_set_v1($1,$2,$3,$4,$5)`,
		recoverySetID, hashB, []byte(`{"protected":true}`), databaseSystemID, now)
}

func requestReadableRestore(
	t *testing.T, ctx context.Context, database *sql.DB,
	source, safety string, now time.Time,
) {
	t.Helper()
	operationID := "99999999-9999-7999-8999-999999999999"
	var previousDatabaseSystemID string
	if err := database.QueryRowContext(ctx, `SELECT current_database_system_id
		FROM platform.platformops_database_lineages WHERE environment_id=$1`, environmentID).
		Scan(&previousDatabaseSystemID); err != nil {
		t.Fatal(err)
	}
	query := `SELECT platform.platformops_request_restore_v1(
		$1,$2,$3,'existing_readable',$4,$5,$6,$7,NULL,NULL,false,$8)`
	if _, err := database.ExecContext(ctx, query, operationID, source, environmentID,
		hashA, []byte(`{"restore":true}`), safety, hashB, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, query,
		"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", source, environmentID,
		hashA, []byte(`same`), source, hashB, now); err == nil {
		t.Fatal("source recovery set was accepted as its own safety set")
	}
	for _, transition := range [][2]string{
		{"plan_ready", "draining"}, {"draining", "restoring"},
		{"restoring", "verifying"}, {"verifying", "switched"},
	} {
		assertBooleanRoutine(t, ctx, database,
			`SELECT platform.platformops_advance_restore_v1($1,$2,$3,$4)`,
			operationID, transition[0], transition[1], now)
	}
	receiptBytes := []byte(`{"completed":true}`)
	receiptDigest := sha256.Sum256(receiptBytes)
	receiptSHA256 := hex.EncodeToString(receiptDigest[:])
	assertBooleanRoutine(t, ctx, database,
		`SELECT platform.platformops_complete_restore_v1($1,$2,$3,$4,$5)`,
		operationID, receiptSHA256, receiptBytes, lineageID, now)
	assertBooleanRoutine(t, ctx, database,
		`SELECT platform.platformops_commit_database_lineage_restore_v1($1,$2,$3,$4)`,
		operationID, previousDatabaseSystemID, lineageID, now)
	var generation int64
	var current string
	if err := database.QueryRowContext(ctx, `SELECT generation,current_database_system_id
		FROM platform.platformops_database_lineages WHERE environment_id=$1`, environmentID).
		Scan(&generation, &current); err != nil {
		t.Fatal(err)
	}
	if generation != 2 || current != lineageID {
		t.Fatalf("lineage generation/system = %d/%s", generation, current)
	}
	var seeded, ownerChanged, tenantReplaced bool
	if err := database.QueryRowContext(ctx, `SELECT seeded, owner_changed, tenant_replaced
		FROM platform.platformops_restore_operations WHERE operation_id=$1`, operationID).
		Scan(&seeded, &ownerChanged, &tenantReplaced); err != nil {
		t.Fatal(err)
	}
	if seeded || ownerChanged || tenantReplaced {
		t.Fatal("restore changed protected installation identity")
	}
	exerciseRestoredLicenseGrant(
		t, ctx, database, operationID, receiptSHA256,
	)
}

func assertBooleanRoutine(
	t *testing.T, ctx context.Context, database *sql.DB, query string, arguments ...any,
) {
	t.Helper()
	var changed bool
	if err := database.QueryRowContext(ctx, query, arguments...).Scan(&changed); err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("lifecycle transition did not change the row")
	}
}
