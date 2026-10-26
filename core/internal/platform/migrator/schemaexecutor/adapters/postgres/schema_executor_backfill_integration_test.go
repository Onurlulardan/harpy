//go:build integration

package postgres

import (
	"context"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func runSchemaExecutorBackfill(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantID string,
) {
	runSchemaExecutorBackfillResume(t, ctx, environment, tenantID)
	runSchemaExecutorBackfillSuppression(t, ctx, environment, tenantID)
}

func runSchemaExecutorBackfillResume(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantID string,
) {
	formID, fieldID := schemaExecutorIntegrationUUID(500), schemaExecutorIntegrationUUID(501)
	planValue := schemaExecutorBackfillPlan(t, tenantID, formID, fieldID)
	createSchemaExecutorBackfillTable(t, ctx, environment, "tenant_a", planValue, 1500)
	jobID := schemaExecutorIntegrationUUID(502)
	environment.seedPlan(t, ctx, jobID, planValue)
	_, lease, plan := environment.startPlan(t, ctx,
		schemaExecutorIntegrationUUID(503), schemaExecutorIntegrationUUID(504))
	session, err := environment.store.openTenantSession(ctx, lease)
	if err != nil {
		t.Fatal(err)
	}
	complete, err := environment.store.applyBackfillBatch(
		ctx, session, schemaexecutor.ApplyRequest{Lease: lease, Plan: plan},
	)
	session.close(environment.store.cleanupTimeout)
	if err != nil || complete {
		t.Fatalf("first backfill batch complete=%t err=%v", complete, err)
	}
	var processed int64
	var checkpoint *string
	if err := environment.pool.QueryRow(ctx, `
SELECT processed_rows, checkpoint_record_id::text
FROM platform.schema_job_steps WHERE job_id = $1`, jobID,
	).Scan(&processed, &checkpoint); err != nil {
		t.Fatal(err)
	}
	if processed != 1000 || checkpoint == nil {
		t.Fatalf("first checkpoint rows=%d checkpoint=%v", processed, checkpoint)
	}
	if err := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{
		Lease: lease, Plan: plan,
	}); err != nil {
		t.Fatalf("resume backfill: %v", err)
	}
	form := planValue.Steps[0].IdentifierMapping.Form.PhysicalName
	fields := planValue.Steps[0].IdentifierMapping.Fields
	query := "SELECT pg_catalog.count(*) FROM " + pgIdentifier("tenant_a") + "." +
		pgIdentifier(form) + " WHERE " + pgIdentifier(fields[0].PhysicalName) +
		" IS DISTINCT FROM " + pgIdentifier(fields[1].PhysicalName)
	var mismatches int
	if err := environment.pool.QueryRow(ctx, query).Scan(&mismatches); err != nil {
		t.Fatal(err)
	}
	if mismatches != 0 {
		t.Fatalf("backfill mismatches = %d", mismatches)
	}
}

func runSchemaExecutorBackfillSuppression(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantID string,
) {
	formID, fieldID := schemaExecutorIntegrationUUID(510), schemaExecutorIntegrationUUID(511)
	planValue := schemaExecutorBackfillPlan(t, tenantID, formID, fieldID)
	createSchemaExecutorBackfillTable(t, ctx, environment, "tenant_a", planValue, 10)
	form := planValue.Steps[0].IdentifierMapping.Form.PhysicalName
	installSchemaExecutorSuppressionTrigger(t, ctx, environment, "tenant_a", form)
	jobID := schemaExecutorIntegrationUUID(512)
	environment.seedPlan(t, ctx, jobID, planValue)
	_, lease, plan := environment.startPlan(t, ctx,
		schemaExecutorIntegrationUUID(513), schemaExecutorIntegrationUUID(514))
	err := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{Lease: lease, Plan: plan})
	if err == nil {
		t.Fatal("expected suppressed backfill rejection")
	}
	kind, code := schemaexecutor.FaultDetails(err)
	if kind != schemaexecutor.FaultBlocked || code != "schema_backfill_write_suppressed" {
		t.Fatalf("suppressed backfill fault kind=%s code=%s err=%v", kind, code, err)
	}
	var processed int64
	if err := environment.pool.QueryRow(ctx,
		"SELECT processed_rows FROM platform.schema_job_steps WHERE job_id = $1", jobID,
	).Scan(&processed); err != nil || processed != 0 {
		t.Fatalf("suppressed progress=%d err=%v", processed, err)
	}
	if err := environment.store.RecordFailure(ctx, schemaexecutor.FailureRequest{
		Lease: lease, State: schemaexecutor.FailureBlocked,
		Code: code, RedactedDetailSHA256: strings.Repeat("a", 64),
	}); err != nil {
		t.Fatalf("finish suppressed job: %v", err)
	}
}
