//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
)

func assertSettingsReplayAndIsolation(t *testing.T, ctx context.Context,
	database *sql.DB, primary, secondary string,
) {
	t.Helper()
	digest := bytes.Repeat([]byte{0x41}, 32)
	first := executeSettingsUpdate(ctx, database, primary, "Primary Updated", 1,
		"tenant-settings-0001", digest)
	if first.err != nil || first.revision != 2 || first.replay || first.outcome != "applied" {
		t.Fatalf("first update = %+v", first)
	}
	replay := executeSettingsUpdate(ctx, database, primary, "Primary Updated", 1,
		"tenant-settings-0001", digest)
	if replay.err != nil || replay.revision != 2 || !replay.replay || replay.outcome != "replay" {
		t.Fatalf("replay = %+v", replay)
	}
	conflict := executeSettingsUpdate(ctx, database, primary, "Different", 1,
		"tenant-settings-0001", bytes.Repeat([]byte{0x42}, 32))
	if conflict.err != nil || conflict.replay || conflict.outcome != "idempotency_conflict" {
		t.Fatalf("idempotency conflict = %+v", conflict)
	}
	isolated := executeSettingsUpdate(ctx, database, secondary, "Secondary Updated", 1,
		"tenant-settings-0001", digest)
	if isolated.err != nil || isolated.revision != 2 || isolated.outcome != "applied" {
		t.Fatalf("cross-tenant key = %+v", isolated)
	}
	var primaryName, secondaryName string
	err := database.QueryRowContext(ctx, `SELECT
(SELECT name FROM platform.tenants WHERE id=$1),
(SELECT name FROM platform.tenants WHERE id=$2)`, primary, secondary).Scan(
		&primaryName, &secondaryName)
	if err != nil || primaryName != "Primary Updated" || secondaryName != "Secondary Updated" {
		t.Fatalf("cross-tenant names=%q/%q err=%v", primaryName, secondaryName, err)
	}
}

func assertSettingsConcurrency(t *testing.T, ctx context.Context,
	database *sql.DB, tenantID string,
) {
	t.Helper()
	start := make(chan struct{})
	results := make(chan settingsRoutineResult, 2)
	var wait sync.WaitGroup
	for index := 0; index < 2; index++ {
		wait.Add(1)
		go func(candidate int) {
			defer wait.Done()
			<-start
			results <- executeSettingsUpdate(ctx, database, tenantID,
				fmt.Sprintf("Concurrent %d", candidate), 2,
				fmt.Sprintf("tenant-concurrent-%04d", candidate),
				bytes.Repeat([]byte{byte(0x50 + candidate)}, 32))
		}(index)
	}
	close(start)
	wait.Wait()
	close(results)
	applied, conflicts := 0, 0
	for result := range results {
		if result.err != nil {
			t.Fatalf("concurrent update: %v", result.err)
		}
		switch result.outcome {
		case "applied":
			applied++
		case "revision_conflict":
			conflicts++
		default:
			t.Fatalf("unexpected concurrent outcome: %+v", result)
		}
	}
	if applied != 1 || conflicts != 1 {
		t.Fatalf("concurrent outcomes applied=%d conflicts=%d", applied, conflicts)
	}
}

func executeSettingsUpdate(ctx context.Context, database *sql.DB,
	tenantID, name string, expected int64, key string, digest []byte,
) settingsRoutineResult {
	var result settingsRoutineResult
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		result.err = err
		return result
	}
	_, err = transaction.ExecContext(ctx, `
SET LOCAL SESSION AUTHORIZATION harpy_v2_runtime;
SET LOCAL ROLE harpy_runtime`)
	if err != nil {
		_ = transaction.Rollback()
		result.err = err
		return result
	}
	result.err = transaction.QueryRowContext(ctx, `SELECT result_revision,replay,outcome
FROM platform.tenancy_settings_update_v1($1,$2,$3,$4,$5)`, tenantID,
		name, expected, key, digest).Scan(&result.revision, &result.replay, &result.outcome)
	if result.err == nil {
		result.err = transaction.Commit()
	} else {
		_ = transaction.Rollback()
	}
	return result
}

func assertSettingsReceiptImmutable(t *testing.T, ctx context.Context,
	database *sql.DB, tenantID string,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `UPDATE platform.tenancy_settings_operation_receipts
SET result_revision=result_revision+1 WHERE tenant_id=$1`, tenantID)
	if err == nil {
		t.Fatal("tenant settings receipt update was not rejected")
	}
}

func assertSettingsV1Fence(t *testing.T, ctx context.Context,
	database *sql.DB, tenantID string,
) {
	t.Helper()
	var before, after int64
	if err := database.QueryRowContext(ctx,
		"SELECT settings_revision FROM platform.tenants WHERE id=$1", tenantID,
	).Scan(&before); err != nil {
		t.Fatal(err)
	}
	_, err := database.ExecContext(ctx, `UPDATE public.tenants
SET name=name||' V1',settings='{"legacy":true}' WHERE id=$1`, tenantID)
	if err != nil {
		t.Fatalf("V1 settings fence update: %v", err)
	}
	if err = database.QueryRowContext(ctx,
		"SELECT settings_revision FROM platform.tenants WHERE id=$1", tenantID,
	).Scan(&after); err != nil || after != before+1 {
		t.Fatalf("V1 settings revision=%d -> %d err=%v", before, after, err)
	}
	result := executeSettingsUpdate(ctx, database, tenantID, "Stale V2", before,
		"tenant-after-v1-0001", bytes.Repeat([]byte{0x61}, 32))
	if result.err != nil || result.outcome != "revision_conflict" || result.revision != after {
		t.Fatalf("stale V2 after V1 update = %+v", result)
	}
}
