//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestRuntimePoolAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	database := startRuntimeTestDatabase(t, ctx)
	pool, err := Open(
		ctx, runtimePoolConfig(t, database.runtimeDSN), ProfileAPI,
		runtimeSchemaContract(t),
	)
	if err != nil {
		t.Fatalf("open runtime pool: %v", err)
	}
	t.Cleanup(pool.Close)
	result := pool.StartupResult()
	if result.ServerVersionNumber() != 180004 || result.VectorSchema() != "platform_ext" ||
		result.DatabaseOrigin() != runtimev1.DatabaseOriginFresh ||
		result.DatabaseName() != "harpy_runtime" || result.SystemIdentifier() == "" ||
		result.BinaryContractDigest() == "" {
		t.Fatalf("unexpected startup evidence: %+v", result)
	}
	t.Run("schema contract", func(t *testing.T) {
		assertSchemaContractProbe(t, ctx, pool, database)
	})
	t.Run("readiness observation", func(t *testing.T) {
		assertReadinessObservation(t, ctx, pool, database)
	})
	t.Run("tainted sessions", func(t *testing.T) {
		assertTaintedSessionsDestroyed(t, ctx, pool)
	})
	t.Run("runtime privileges", func(t *testing.T) {
		assertRuntimePrivileges(t, ctx, pool)
	})
	t.Run("tenant slug locator", func(t *testing.T) {
		assertTenantSlugLocator(t, ctx, pool, database)
	})
	t.Run("bound tenant schema", func(t *testing.T) {
		assertBoundCurrentSchema(t, ctx, pool, database)
	})
	t.Run("foundation mismatch", func(t *testing.T) {
		assertFoundationMismatchRejected(t, database)
		assertFoundationRunCommandRejected(t, database)
	})
	t.Run("readiness drift", func(t *testing.T) {
		assertReadinessDriftDetection(t, ctx, pool, database)
	})
	t.Run("bounded force close", func(t *testing.T) {
		assertForceCloseDoesNotWaitForBorrowedConnection(t, ctx, pool)
	})
}

func assertForceCloseDoesNotWaitForBorrowedConnection(
	t *testing.T,
	ctx context.Context,
	pool *Pool,
) {
	t.Helper()
	connection, err := pool.inner.Acquire(ctx)
	if err != nil {
		t.Fatalf("borrow connection: %v", err)
	}
	started := time.Now()
	pool.ForceClose()
	if elapsed := time.Since(started); elapsed > time.Second {
		connection.Release()
		t.Fatalf("force close blocked for %s", elapsed)
	}
	if err := pool.Ping(ctx); err != ErrPoolClosed {
		connection.Release()
		t.Fatalf("closed pool ping error=%v", err)
	}
	pool.ForceClose()
	connection.Release()
}
