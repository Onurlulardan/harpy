//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func assertIAMPlatformGuardRaces(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	assertIAMThrottleConcurrency(t, ctx, database)
	assertIAMLocatorDoubleAdvance(t, ctx, database)
	assertIAMAuthoritySwitchRace(t, ctx, database)
	assertIAMLoginResolverRace(t, ctx, database)
}

func assertIAMAuthoritySwitchRace(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	now := time.Now().UTC()
	runtimeTransaction, err := database.runtime.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = createIAMGuardLocator(ctx, runtimeTransaction, iamBytes(81, 16),
		iamGuardTenantB, 1, "active", now.Add(-time.Minute), now,
		now.Add(48*time.Hour), now.Add(73*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	assertIAMSQLState(t, tryIAMGroupSwitch(ctx, database.migrator,
		iamGuardTenantB, 2), "55P03")
	if err := runtimeTransaction.Commit(); err != nil {
		t.Fatal(err)
	}
	advanceIAMGuardAuthority(t, ctx, database.migrator, iamGuardTenantB, 2)

	switchTransaction, err := database.migrator.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := updateIAMGuardAuthority(ctx, switchTransaction, iamGuardTenantB, 3); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		done <- createIAMGuardLocator(ctx, database.runtime, iamBytes(82, 16),
			iamGuardTenantB, 2, "active", now.Add(-time.Minute), now,
			now.Add(48*time.Hour), now.Add(73*time.Hour))
	}()
	assertIAMOperationBlocks(t, done)
	if err := switchTransaction.Commit(); err != nil {
		t.Fatal(err)
	}
	assertIAMSQLState(t, awaitIAMGuardResult(t, done), "P5701")
}

func tryIAMGroupSwitch(
	ctx context.Context, database *sql.DB, tenantID string, generation int64,
) error {
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	if _, err := transaction.ExecContext(ctx, "SET LOCAL lock_timeout='150ms'"); err != nil {
		return err
	}
	_, err = transaction.ExecContext(ctx, `UPDATE platform.cutover_groups
SET generation=$2 WHERE scope_kind='tenant' AND scope_id=$1
AND group_id='iam_system_resources'`, tenantID, generation)
	return err
}

func updateIAMGuardAuthority(
	ctx context.Context, transaction *sql.Tx, tenantID string, generation int64,
) error {
	if _, err := transaction.ExecContext(ctx, `UPDATE platform.cutover_groups
SET generation=$2 WHERE scope_kind='tenant' AND scope_id=$1
AND group_id='iam_system_resources'`, tenantID, generation); err != nil {
		return err
	}
	_, err := transaction.ExecContext(ctx, `UPDATE platform.domain_cutovers
SET generation=$2 WHERE scope_kind='tenant' AND scope_id=$1
AND domain IN ('iam','system_resource_layout')`, tenantID, generation)
	return err
}

func assertIAMOperationBlocks(t *testing.T, done <-chan error) {
	t.Helper()
	select {
	case err := <-done:
		t.Fatalf("operation did not block: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
}

func awaitIAMGuardResult(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for IAM guard race")
		return nil
	}
}
