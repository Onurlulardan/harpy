//go:build integration

package postgres

import (
	"context"
	"testing"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func (environment *schemaExecutorIntegrationEnvironment) seedPlanV2(
	t *testing.T,
	ctx context.Context,
	jobID string,
	plan schemav2.Plan,
) schemaexecutor.PreparedPlan {
	t.Helper()
	prepared, err := schemaexecutor.BuildPlanV2(plan)
	if err != nil {
		t.Fatalf("build V2 integration plan: %v", err)
	}
	transaction, err := environment.pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin V2 schema plan insert: %v", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()
	_, err = transaction.Exec(ctx, `
INSERT INTO platform.schema_jobs (
    job_id, tenant_id, owner, resource_kind, resource_id, resource_revision,
    compiler_version, plan_contract_version, canonical_plan, canonical_plan_bytes,
    plan_sha256, step_count, available_at, submission_authority_id,
    submission_authority_group_id, submission_group_generation,
    submission_billing_generation, submission_forms_generation
) VALUES ($1, $2, 'forms', $3, $4, $5, $6, $7, $8::jsonb, $9, $10, $11,
    pg_catalog.clock_timestamp(), $12, 'billing_forms', $13, $14, $15)`,
		jobID, plan.TenantID, string(plan.Resource.Kind), plan.Resource.ID,
		plan.Resource.Revision, plan.Compiler.Version, plan.ContractVersion,
		string(prepared.CanonicalJSON()), prepared.CanonicalJSON(), prepared.SHA256(),
		len(plan.Steps), v2ExactTarget, schemaSubmissionV2GroupGeneration,
		schemaSubmissionV2BillingGeneration, schemaSubmissionV2FormsGeneration,
	)
	if err != nil {
		t.Fatalf("insert V2 schema job: %v", err)
	}
	for index, step := range plan.Steps {
		stored := prepared.Steps()[index]
		_, err = transaction.Exec(ctx, `
INSERT INTO platform.schema_job_steps (
    job_id, ordinal, plan_sha256, operation_id, transaction_class, executor_kind,
    identifier_mapping, arguments, canonical_step_bytes, step_sha256, next_executor
) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8::jsonb, $9, $10, $6)`,
			jobID, index, prepared.SHA256(), string(step.OperationID),
			string(step.TransactionClass), string(step.ExecutorKind),
			string(stored.IdentifierMappingJSON()), string(stored.ArgumentsJSON()),
			stored.CanonicalPreimage(), stored.SHA256(),
		)
		if err != nil {
			t.Fatalf("insert V2 schema step %d: %v", index, err)
		}
	}
	if err := transaction.Commit(ctx); err != nil {
		t.Fatalf("commit V2 schema plan insert: %v", err)
	}
	return prepared
}
