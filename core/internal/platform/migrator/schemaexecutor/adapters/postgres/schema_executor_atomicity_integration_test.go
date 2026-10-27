//go:build integration

package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func runSchemaExecutorTransactionAtomicity(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantID string,
) {
	formID, jobID := schemaExecutorIntegrationUUID(400), schemaExecutorIntegrationUUID(401)
	planValue := schemaExecutorCreatePlan(t, tenantID, formID, 1)
	environment.seedPlan(t, ctx, jobID, planValue)
	_, lease, plan := environment.startPlan(t, ctx,
		schemaExecutorIntegrationUUID(402), schemaExecutorIntegrationUUID(403))
	installSchemaExecutorCompletionFailure(t, ctx, environment, jobID)
	err := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{Lease: lease, Plan: plan})
	if err == nil {
		t.Fatal("expected completion trigger failure")
	}
	formName := planValue.Steps[0].IdentifierMapping.Form.PhysicalName
	var relation *string
	if err := environment.pool.QueryRow(ctx,
		"SELECT pg_catalog.to_regclass($1)::text", "tenant_a."+formName,
	).Scan(&relation); err != nil {
		t.Fatal(err)
	}
	if relation != nil {
		t.Fatalf("DDL survived failed state transaction: %s", *relation)
	}
	removeSchemaExecutorCompletionFailure(t, ctx, environment)
	if err := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{
		Lease: lease, Plan: plan,
	}); err != nil {
		t.Fatalf("apply atomic schema step: %v", err)
	}
	var jobState, stepState string
	var receipts int
	if err := environment.pool.QueryRow(ctx, `
SELECT job.state, step.state,
       (SELECT pg_catalog.count(*)::integer
        FROM platform.schema_job_step_attempts AS attempt
        WHERE attempt.job_id = job.job_id)
FROM platform.schema_jobs AS job
JOIN platform.schema_job_steps AS step ON step.job_id = job.job_id
WHERE job.job_id = $1`, jobID).Scan(&jobState, &stepState, &receipts); err != nil {
		t.Fatal(err)
	}
	if jobState != "succeeded" || stepState != "succeeded" || receipts != 1 {
		t.Fatalf("atomic outcome job=%s step=%s receipts=%d", jobState, stepState, receipts)
	}
}

func installSchemaExecutorCompletionFailure(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	jobID string,
) {
	t.Helper()
	_, err := environment.pool.Exec(ctx, `
CREATE FUNCTION platform.schema_executor_test_reject_completion()
RETURNS trigger LANGUAGE plpgsql SET search_path TO pg_catalog AS $function$
BEGIN
    RAISE EXCEPTION 'test completion rejection';
END;
$function$`)
	if err != nil {
		t.Fatal(err)
	}
	statement := fmt.Sprintf(`
CREATE TRIGGER aaa_schema_executor_test_reject_completion
BEFORE UPDATE ON platform.schema_job_steps
FOR EACH ROW WHEN (NEW.job_id = '%s'::uuid AND NEW.state = 'succeeded')
EXECUTE FUNCTION platform.schema_executor_test_reject_completion()`, jobID)
	if _, err := environment.pool.Exec(ctx, statement); err != nil {
		t.Fatal(err)
	}
}

func removeSchemaExecutorCompletionFailure(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
) {
	t.Helper()
	if _, err := environment.pool.Exec(ctx, `
DROP TRIGGER aaa_schema_executor_test_reject_completion
ON platform.schema_job_steps`); err != nil {
		t.Fatal(err)
	}
	if _, err := environment.pool.Exec(ctx,
		"DROP FUNCTION platform.schema_executor_test_reject_completion()",
	); err != nil {
		t.Fatal(err)
	}
}
