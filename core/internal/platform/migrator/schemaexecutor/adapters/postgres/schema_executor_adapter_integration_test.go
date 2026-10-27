//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestSchemaExecutorAdapterOnPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	environment := startSchemaExecutorIntegration(t, ctx)
	tenantA, tenantB := schemaExecutorIntegrationUUID(100), schemaExecutorIntegrationUUID(101)
	environment.createTenant(t, ctx, tenantA, "tenant_a")
	environment.createTenant(t, ctx, tenantB, "tenant_b")
	t.Run("claim race", func(t *testing.T) {
		runSchemaExecutorClaimRace(t, ctx, environment, tenantA)
	})
	t.Run("multi step attempts", func(t *testing.T) {
		runSchemaExecutorMultiStepAttempts(t, ctx, environment, tenantA)
	})
	t.Run("stale heartbeat and recovery", func(t *testing.T) {
		runSchemaExecutorLeaseRecovery(t, ctx, environment, tenantA)
	})
	t.Run("transaction rollback and commit", func(t *testing.T) {
		runSchemaExecutorTransactionAtomicity(t, ctx, environment, tenantA)
	})
	t.Run("backfill resume and suppression", func(t *testing.T) {
		runSchemaExecutorBackfill(t, ctx, environment, tenantA)
	})
	t.Run("concurrent index reconcile", func(t *testing.T) {
		runSchemaExecutorConcurrentIndexes(t, ctx, environment, tenantA)
	})
	t.Run("tenant isolation and advisory lock", func(t *testing.T) {
		runSchemaExecutorTenantIsolation(t, ctx, environment, tenantA, tenantB)
	})
	t.Run("constraint catalog drift", func(t *testing.T) {
		runSchemaExecutorConstraintCatalog(t, ctx, environment, tenantA)
	})
	t.Run("tenant verification transport error", func(t *testing.T) {
		runSchemaExecutorVerifyTenantTransportError(t, ctx, environment, tenantA)
	})
	t.Run("bounded row lock", func(t *testing.T) {
		runSchemaExecutorBoundedRowLock(t, ctx, environment, tenantA)
	})
	assertSchemaExecutorNoActiveJobs(t, ctx, environment)
}
