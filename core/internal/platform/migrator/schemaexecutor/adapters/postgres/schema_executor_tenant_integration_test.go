//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func runSchemaExecutorTenantIsolation(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantA string,
	tenantB string,
) {
	t.Helper()
	runSchemaExecutorCrossTenantIsolation(t, ctx, environment, tenantA)
	runSchemaExecutorSentinelMismatch(t, ctx, environment, tenantA, tenantB)
	runSchemaExecutorOwnerMismatch(t, ctx, environment, tenantB)
	runSchemaExecutorAdvisoryLockWait(t, ctx, environment, tenantA)
}

func runSchemaExecutorCrossTenantIsolation(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantID string,
) {
	formID, jobID := schemaExecutorIntegrationUUID(700), schemaExecutorIntegrationUUID(701)
	planValue := schemaExecutorCreatePlan(t, tenantID, formID, 1)
	environment.seedPlan(t, ctx, jobID, planValue)
	_, lease, plan := environment.startPlan(t, ctx,
		schemaExecutorIntegrationUUID(702), schemaExecutorIntegrationUUID(703))
	if err := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{
		Lease: lease, Plan: plan,
	}); err != nil {
		t.Fatalf("apply isolated tenant plan: %v", err)
	}
	name := planValue.Steps[0].IdentifierMapping.Form.PhysicalName
	assertSchemaExecutorRelationExists(t, ctx, environment, "tenant_a", name, true)
	assertSchemaExecutorRelationExists(t, ctx, environment, "tenant_b", name, false)
}

func runSchemaExecutorSentinelMismatch(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	wrongTenantID string,
	tenantID string,
) {
	formID, jobID := schemaExecutorIntegrationUUID(710), schemaExecutorIntegrationUUID(711)
	environment.seedPlan(t, ctx, jobID, schemaExecutorCreatePlan(t, tenantID, formID, 1))
	_, lease, plan := environment.startPlan(t, ctx,
		schemaExecutorIntegrationUUID(712), schemaExecutorIntegrationUUID(713))
	if _, err := environment.pool.Exec(ctx,
		`UPDATE tenant_b.tenancy_schema_identity SET tenant_id = $1`, wrongTenantID,
	); err != nil {
		t.Fatal(err)
	}
	applyErr := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{Lease: lease, Plan: plan})
	if _, err := environment.pool.Exec(ctx,
		`UPDATE tenant_b.tenancy_schema_identity SET tenant_id = $1`, tenantID,
	); err != nil {
		t.Fatal(err)
	}
	schemaExecutorAssertFault(t, applyErr,
		schemaexecutor.FaultBlocked, "schema_tenant_sentinel_invalid")
	recordSchemaExecutorBlocked(t, ctx, environment, lease, "schema_tenant_sentinel_invalid")
}

func runSchemaExecutorOwnerMismatch(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantID string,
) {
	formID, jobID := schemaExecutorIntegrationUUID(720), schemaExecutorIntegrationUUID(721)
	environment.seedPlan(t, ctx, jobID, schemaExecutorCreatePlan(t, tenantID, formID, 1))
	_, lease, plan := environment.startPlan(t, ctx,
		schemaExecutorIntegrationUUID(722), schemaExecutorIntegrationUUID(723))
	if _, err := environment.admin.Exec(ctx,
		`ALTER SCHEMA tenant_b OWNER TO harpy_wrong_owner`); err != nil {
		t.Fatal(err)
	}
	applyErr := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{Lease: lease, Plan: plan})
	if _, err := environment.admin.Exec(ctx,
		`ALTER SCHEMA tenant_b OWNER TO harpy_owner`); err != nil {
		t.Fatal(err)
	}
	schemaExecutorAssertFault(t, applyErr,
		schemaexecutor.FaultBlocked, "schema_tenant_invalid")
	recordSchemaExecutorBlocked(t, ctx, environment, lease, "schema_tenant_invalid")
}

func runSchemaExecutorAdvisoryLockWait(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantID string,
) {
	formID, jobID := schemaExecutorIntegrationUUID(730), schemaExecutorIntegrationUUID(731)
	environment.seedPlan(t, ctx, jobID, schemaExecutorCreatePlan(t, tenantID, formID, 1))
	_, lease, plan := environment.startPlan(t, ctx,
		schemaExecutorIntegrationUUID(732), schemaExecutorIntegrationUUID(733))
	connection, err := environment.pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Release()
	const lockSQL = `SELECT pg_catalog.pg_advisory_lock(pg_catalog.hashtextextended(
'harpy:v2:migrate:tenant:' || $1, 0))`
	if _, err := connection.Exec(ctx, lockSQL, tenantID); err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	started := time.Now()
	go func() {
		result <- environment.store.Apply(ctx, schemaexecutor.ApplyRequest{Lease: lease, Plan: plan})
	}()
	timer := time.NewTimer(schemaexecutor.DatabaseOperationTimeout + time.Second)
	defer timer.Stop()
	select {
	case err := <-result:
		t.Fatalf("tenant lock wait ended early: %v", err)
	case <-timer.C:
	}
	var released bool
	if err := connection.QueryRow(ctx,
		`SELECT pg_catalog.pg_advisory_unlock(pg_catalog.hashtextextended(
'harpy:v2:migrate:tenant:' || $1, 0))`, tenantID,
	).Scan(&released); err != nil || !released {
		t.Fatalf("release held tenant lock: released=%t err=%v", released, err)
	}
	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("apply after healthy tenant lock wait: %v", err)
		}
	case <-time.After(schemaexecutor.DatabaseOperationTimeout):
		t.Fatal("apply did not resume after tenant lock release")
	}
	if time.Since(started) <= schemaexecutor.DatabaseOperationTimeout {
		t.Fatal("tenant lock wait did not outlive the bounded DB operation timeout")
	}
}
