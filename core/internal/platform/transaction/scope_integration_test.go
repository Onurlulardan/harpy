//go:build integration

package transaction

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func TestTenantScopeAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := scopeTestContext(t)
	defer cancel()
	database := startScopeTestDatabase(t, ctx)
	pool := openScopePool(t, ctx, database.runtimeDSN)
	releaseReservation, err := postgres.ReserveIntegrationConnection(ctx, pool)
	if err != nil {
		t.Fatal(err)
	}
	defer releaseReservation()
	coordinator, err := NewCoordinator(pool)
	if err != nil {
		t.Fatal(err)
	}
	binder := scopeProbeBinder(t)
	t.Run("commit and tenant isolation", func(t *testing.T) {
		assertScopeCommitIsolation(t, ctx, coordinator, binder)
	})
	t.Run("rollback read-only cancel and panic", func(t *testing.T) {
		assertScopeRollbackPaths(t, ctx, coordinator, binder)
	})
	t.Run("lifetime and nested scope", func(t *testing.T) {
		assertScopeLifetime(t, ctx, coordinator, binder)
	})
	t.Run("lazy row cannot escape", func(t *testing.T) {
		assertUnscannedRowNeverEscapes(t, ctx, coordinator, binder, database)
	})
	t.Run("abandoned rows rollback and destroy", func(t *testing.T) {
		assertAbandonedRowsRollbackAndDestroy(t, ctx, coordinator, binder, database)
	})
	t.Run("queued executor is rejected at seal", func(t *testing.T) {
		assertQueuedExecutorRejectedAtSeal(t, ctx, coordinator, binder, database)
	})
	t.Run("cancellation cleans active rows", func(t *testing.T) {
		assertCancellationCleansActiveRows(t, ctx, coordinator, binder)
	})
	t.Run("tenant and authority denials", func(t *testing.T) {
		assertScopeDenials(t, ctx, coordinator, binder, database)
	})
	t.Run("sealed owner authority lifecycle", func(t *testing.T) {
		assertScopeOwnerAuthority(t, ctx, coordinator)
	})
	t.Run("multi-group command authority", func(t *testing.T) {
		assertMultiGroupCommandAuthority(t, ctx, coordinator, database)
	})
	t.Run("credential-only tenant scope", func(t *testing.T) {
		assertCredentialTenantScope(t, ctx, coordinator)
	})
	t.Run("credential target substitution", func(t *testing.T) {
		assertCredentialTargetSubstitutionRejected(t, ctx, coordinator)
	})
	t.Run("job tenant identity binding", func(t *testing.T) {
		assertJobTenantScopeBindsIdentityBeforeOwnerQueries(t, ctx, coordinator)
	})
	t.Run("exact multi-owner snapshot authority", func(t *testing.T) {
		assertMultiOwnerSnapshotAuthority(t, ctx, coordinator)
	})
	t.Run("stale snapshot authority", func(t *testing.T) {
		assertStaleSnapshotAuthority(t, ctx, coordinator)
	})
	t.Run("global command identity search path and commit", func(t *testing.T) {
		assertGlobalScopeCommit(t, ctx, coordinator, database)
	})
	t.Run("public global catalog snapshot", func(t *testing.T) {
		assertGlobalPublicSnapshot(t, ctx, coordinator)
	})
	t.Run("global authority and identity denials", func(t *testing.T) {
		assertGlobalScopeDenials(t, ctx, coordinator, database)
	})
	t.Run("global and tenant nesting", func(t *testing.T) {
		assertGlobalScopeNesting(t, ctx, coordinator)
	})
	t.Run("ambiguous commit", func(t *testing.T) {
		assertAmbiguousCommit(t, ctx, coordinator, binder, database)
	})
}
