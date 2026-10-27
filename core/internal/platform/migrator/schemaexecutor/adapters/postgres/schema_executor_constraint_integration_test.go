//go:build integration

package postgres

import (
	"context"
	"testing"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func runSchemaExecutorConstraintCatalog(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantID string,
) {
	t.Helper()
	runSchemaExecutorConstraintDrift(t, ctx, environment, tenantID)
	runSchemaExecutorConstraintValidation(t, ctx, environment, tenantID)
}

func runSchemaExecutorConstraintDrift(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantID string,
) {
	planValue := schemaExecutorConstraintPlan(t, tenantID,
		schemaExecutorIntegrationUUID(800), schemaExecutorIntegrationUUID(801),
		schemaExecutorIntegrationUUID(802))
	createSchemaExecutorConstraintTable(t, ctx, environment, planValue, false)
	environment.seedPlan(t, ctx, schemaExecutorIntegrationUUID(803), planValue)
	_, lease, plan := environment.startPlan(t, ctx,
		schemaExecutorIntegrationUUID(804), schemaExecutorIntegrationUUID(805))
	err := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{Lease: lease, Plan: plan})
	schemaExecutorAssertFault(t, err,
		schemaexecutor.FaultBlocked, "schema_constraint_contract_mismatch")
	recordSchemaExecutorBlocked(t, ctx, environment, lease,
		"schema_constraint_contract_mismatch")
}

func runSchemaExecutorConstraintValidation(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantID string,
) {
	planValue := schemaExecutorConstraintPlan(t, tenantID,
		schemaExecutorIntegrationUUID(810), schemaExecutorIntegrationUUID(811),
		schemaExecutorIntegrationUUID(812))
	createSchemaExecutorConstraintTable(t, ctx, environment, planValue, true)
	environment.seedPlan(t, ctx, schemaExecutorIntegrationUUID(813), planValue)
	_, lease, plan := environment.startPlan(t, ctx,
		schemaExecutorIntegrationUUID(814), schemaExecutorIntegrationUUID(815))
	if err := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{
		Lease: lease, Plan: plan,
	}); err != nil {
		t.Fatalf("validate exact constraint: %v", err)
	}
	mapping := planValue.Steps[0].IdentifierMapping
	contract, exists, err := schemaExecutorReadConstraint(ctx, environment.pool,
		"tenant_a", mapping.Form.PhysicalName, mapping.Constraints[0].PhysicalName)
	if err != nil || !exists || !contract.validated {
		t.Fatalf("validated constraint exists=%t contract=%#v err=%v",
			exists, contract, err)
	}
}

func createSchemaExecutorConstraintTable(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	planValue schemav1.Plan,
	exact bool,
) {
	t.Helper()
	mapping := planValue.Steps[0].IdentifierMapping
	table := pgIdentifier("tenant_a") + "." + pgIdentifier(mapping.Form.PhysicalName)
	field := pgIdentifier(mapping.Fields[0].PhysicalName)
	statement := "CREATE TABLE " + table + ` (
    id uuid PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT pg_catalog.statement_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT pg_catalog.statement_timestamp(),
    row_version bigint NOT NULL DEFAULT 1 CHECK (row_version > 0),
    deleted_at timestamptz,
    ` + field + ` text
)`
	if _, err := environment.pool.Exec(ctx, statement); err != nil {
		t.Fatal(err)
	}
	expression := field + " <> ''"
	if exact {
		expression = field + " IS NOT NULL"
	}
	statement = "ALTER TABLE " + table + " ADD CONSTRAINT " +
		pgIdentifier(mapping.Constraints[0].PhysicalName) + " CHECK (" + expression + ") NOT VALID"
	if _, err := environment.pool.Exec(ctx, statement); err != nil {
		t.Fatal(err)
	}
}
