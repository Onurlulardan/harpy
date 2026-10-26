//go:build integration

package postgres

import (
	"context"
	"encoding/base32"
	"fmt"
	"strconv"
	"strings"
	"testing"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/google/uuid"
)

func schemaExecutorIntegrationUUID(value int) string {
	return fmt.Sprintf("00000000-0000-4000-8000-%012d", value)
}

func schemaExecutorObjectName(t *testing.T, prefix, logicalID string) string {
	t.Helper()
	parsed, err := uuid.Parse(logicalID)
	if err != nil {
		t.Fatal(err)
	}
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(parsed[:])
	return prefix + strings.ToLower(encoded)
}

func schemaExecutorFieldNameForTest(
	t *testing.T,
	logicalID string,
	generation int64,
) string {
	t.Helper()
	return schemaExecutorObjectName(t, "fld_", logicalID) +
		"_g" + strconv.FormatInt(generation, 36)
}

func (environment *schemaExecutorIntegrationEnvironment) createTenant(
	t *testing.T,
	ctx context.Context,
	tenantID string,
	schemaName string,
) {
	t.Helper()
	_, err := environment.pool.Exec(ctx, `
INSERT INTO platform.tenants (
    id, name, slug, status, settings, owner_id, schema_name, created_at, updated_at
) VALUES ($1, $2, $3, 'ACTIVE', '{}'::jsonb, $4, $5,
    pg_catalog.clock_timestamp(), pg_catalog.clock_timestamp())`,
		tenantID, schemaName, schemaName, schemaExecutorIntegrationUUID(999), schemaName,
	)
	if err != nil {
		t.Fatalf("insert tenant %s: %v", tenantID, err)
	}
	qualified := pgIdentifier(schemaName)
	if _, err = environment.pool.Exec(ctx, "CREATE SCHEMA "+qualified); err != nil {
		t.Fatalf("create tenant schema %s: %v", schemaName, err)
	}
	_, err = environment.pool.Exec(ctx, `CREATE TABLE `+qualified+`.tenancy_schema_identity (
    singleton boolean PRIMARY KEY CHECK (singleton),
    tenant_id uuid NOT NULL,
    schema_name text NOT NULL
)`)
	if err != nil {
		t.Fatalf("create tenant sentinel %s: %v", schemaName, err)
	}
	_, err = environment.pool.Exec(ctx,
		"INSERT INTO "+qualified+".tenancy_schema_identity VALUES (TRUE, $1, $2)",
		tenantID, schemaName,
	)
	if err != nil {
		t.Fatalf("create tenant schema %s: %v", schemaName, err)
	}
}

func pgIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func schemaExecutorMapping(
	t *testing.T,
	formID string,
	fields []schemav1.FieldIdentifier,
	indexes []schemav1.Identifier,
	constraints []schemav1.Identifier,
) schemav1.IdentifierMapping {
	t.Helper()
	return schemav1.IdentifierMapping{
		Form: schemav1.Identifier{
			LogicalID: formID, PhysicalName: schemaExecutorObjectName(t, "frm_", formID),
		},
		RelatedForms: []schemav1.Identifier{}, Fields: fields,
		Relations: []schemav1.Identifier{}, Indexes: indexes, Constraints: constraints,
	}
}

func schemaExecutorPlan(
	t *testing.T,
	tenantID string,
	formID string,
	revision int64,
	allowed []schemav1.OperationID,
	steps []schemav1.Step,
) schemav1.Plan {
	t.Helper()
	return schemav1.Plan{
		ContractVersion: schemav1.PlanContractVersion,
		TenantID:        tenantID, Owner: "forms",
		Resource: schemav1.Resource{
			Kind: schemav1.ResourceFormStorage, ID: formID, Revision: revision,
		},
		Compiler: schemav1.Compiler{
			Version:         schemav1.InitialCompilerVersion,
			PostgreSQLMajor: schemav1.RequiredPostgreSQLMajor,
		},
		AllowedOperations: allowed, Steps: steps,
	}
}

func (environment *schemaExecutorIntegrationEnvironment) seedPlan(
	t *testing.T,
	ctx context.Context,
	jobID string,
	plan schemav1.Plan,
) schemaexecutor.PreparedPlan {
	t.Helper()
	prepared, err := schemaexecutor.BuildPlan(plan)
	if err != nil {
		t.Fatalf("build integration plan: %v", err)
	}
	transaction, err := environment.pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin schema plan insert: %v", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()
	_, err = transaction.Exec(ctx, `
INSERT INTO platform.schema_jobs (
    job_id, tenant_id, owner, resource_kind, resource_id, resource_revision,
    compiler_version, plan_contract_version, canonical_plan, plan_sha256,
    step_count, available_at
) VALUES ($1, $2, 'forms', $3, $4, $5, $6, $7, $8::jsonb, $9, $10,
    pg_catalog.clock_timestamp())`,
		jobID, plan.TenantID, string(plan.Resource.Kind), plan.Resource.ID,
		plan.Resource.Revision, plan.Compiler.Version, plan.ContractVersion,
		string(prepared.CanonicalJSON()), prepared.SHA256(), len(plan.Steps),
	)
	if err != nil {
		t.Fatalf("insert schema job: %v", err)
	}
	preparedSteps := prepared.Steps()
	for index, step := range plan.Steps {
		_, err = transaction.Exec(ctx, `
INSERT INTO platform.schema_job_steps (
    job_id, ordinal, plan_sha256, operation_id, transaction_class, executor_kind,
    identifier_mapping, arguments, step_sha256, next_executor
) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8::jsonb, $9, $6)`,
			jobID, index, prepared.SHA256(), string(step.OperationID),
			string(step.TransactionClass), string(step.ExecutorKind),
			string(preparedSteps[index].IdentifierMappingJSON()),
			string(preparedSteps[index].ArgumentsJSON()), preparedSteps[index].SHA256(),
		)
		if err != nil {
			t.Fatalf("insert schema step %d: %v", index, err)
		}
	}
	if err := transaction.Commit(ctx); err != nil {
		t.Fatalf("commit schema plan insert: %v", err)
	}
	return prepared
}
