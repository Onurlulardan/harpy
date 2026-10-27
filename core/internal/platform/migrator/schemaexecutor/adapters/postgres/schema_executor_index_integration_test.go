//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func runSchemaExecutorConcurrentIndexes(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantID string,
) {
	runSchemaExecutorValidConcurrentIndex(t, ctx, environment, tenantID, false)
	runSchemaExecutorValidConcurrentIndex(t, ctx, environment, tenantID, true)
	runSchemaExecutorInvalidConcurrentIndex(t, ctx, environment, tenantID)
}

func runSchemaExecutorValidConcurrentIndex(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantID string,
	precreated bool,
) {
	offset := 0
	if precreated {
		offset = 10
	}
	formID := schemaExecutorIntegrationUUID(600 + offset)
	fieldID := schemaExecutorIntegrationUUID(601 + offset)
	indexID := schemaExecutorIntegrationUUID(602 + offset)
	planValue := schemaExecutorIndexPlan(t, tenantID, formID, fieldID, indexID, false)
	createSchemaExecutorIndexTable(t, ctx, environment, "tenant_a", planValue, false)
	if precreated {
		createSchemaExecutorExactIndex(t, ctx, environment, "tenant_a", planValue, false)
	}
	environment.seedPlan(t, ctx, schemaExecutorIntegrationUUID(603+offset), planValue)
	_, lease, plan := environment.startPlan(t, ctx,
		schemaExecutorIntegrationUUID(604+offset), schemaExecutorIntegrationUUID(605+offset))
	if err := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{
		Lease: lease, Plan: plan,
	}); err != nil {
		t.Fatalf("apply valid concurrent index precreated=%t: %v", precreated, err)
	}
	assertSchemaExecutorIndexValid(t, ctx, environment, "tenant_a", planValue)
}

func runSchemaExecutorInvalidConcurrentIndex(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantID string,
) {
	formID := schemaExecutorIntegrationUUID(620)
	fieldID := schemaExecutorIntegrationUUID(621)
	indexID := schemaExecutorIntegrationUUID(622)
	planValue := schemaExecutorIndexPlan(t, tenantID, formID, fieldID, indexID, true)
	createSchemaExecutorIndexTable(t, ctx, environment, "tenant_a", planValue, true)
	jobID := schemaExecutorIntegrationUUID(623)
	environment.seedPlan(t, ctx, jobID, planValue)
	_, lease, plan := environment.startPlan(t, ctx,
		schemaExecutorIntegrationUUID(624), schemaExecutorIntegrationUUID(625))
	createSchemaExecutorExactIndex(t, ctx, environment, "tenant_a", planValue, true)
	err := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{Lease: lease, Plan: plan})
	kind, code := schemaexecutor.FaultDetails(err)
	if kind != schemaexecutor.FaultTransient || code != "schema_invalid_index_reconciled" {
		t.Fatalf("invalid index reconcile kind=%s code=%s err=%v", kind, code, err)
	}
	if err := environment.store.RecordFailure(ctx, schemaexecutor.FailureRequest{
		Lease: lease, State: schemaexecutor.FailureRetryable,
		Code: code, RedactedDetailSHA256: schemaExecutorIntegrationDetailSHA256,
	}); err != nil {
		t.Fatalf("record invalid index retry: %v", err)
	}
	mapping := planValue.Steps[0].IdentifierMapping
	contract, exists, err := schemaExecutorReadIndex(
		ctx, environment.pool, "tenant_a", mapping.Indexes[0].PhysicalName,
	)
	if err != nil || exists {
		t.Fatalf("invalid index survived: exists=%t contract=%#v err=%v", exists, contract, err)
	}
	table := pgIdentifier("tenant_a") + "." + pgIdentifier(mapping.Form.PhysicalName)
	field := pgIdentifier(mapping.Fields[0].PhysicalName)
	if _, err := environment.pool.Exec(ctx,
		"UPDATE "+table+" SET "+field+" = id::text",
	); err != nil {
		t.Fatal(err)
	}
	record, retryLease, retryPlan := environment.startPlan(t, ctx,
		schemaExecutorIntegrationUUID(626), schemaExecutorIntegrationUUID(627))
	if record.Job.JobID != jobID || record.Lease.JobAttempt != 2 ||
		record.Lease.StepAttempt != 2 {
		t.Fatalf("index retry claim job=%s job_attempt=%d step_attempt=%d",
			record.Job.JobID, record.Lease.JobAttempt, record.Lease.StepAttempt)
	}
	if err := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{
		Lease: retryLease, Plan: retryPlan,
	}); err != nil {
		t.Fatalf("retry reconciled index: %v", err)
	}
	assertSchemaExecutorIndexValid(t, ctx, environment, "tenant_a", planValue)
}
