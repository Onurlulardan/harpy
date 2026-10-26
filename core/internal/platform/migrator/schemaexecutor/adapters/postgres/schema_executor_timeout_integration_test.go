//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func runSchemaExecutorBoundedRowLock(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantID string,
) {
	t.Helper()
	formID, jobID := schemaExecutorIntegrationUUID(740), schemaExecutorIntegrationUUID(741)
	environment.seedPlan(t, ctx, jobID, schemaExecutorCreatePlan(t, tenantID, formID, 1))
	record, lease, _ := environment.startPlan(t, ctx,
		schemaExecutorIntegrationUUID(742), schemaExecutorIntegrationUUID(743))
	if record.Job.JobID != jobID {
		t.Fatalf("timeout test claimed job %s, want %s", record.Job.JobID, jobID)
	}
	transaction, err := environment.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()
	var lockedJobID string
	if err := transaction.QueryRow(ctx,
		`SELECT job_id::text FROM platform.schema_jobs WHERE job_id = $1 FOR UPDATE`, jobID,
	).Scan(&lockedJobID); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	err = environment.store.Heartbeat(ctx, schemaexecutor.HeartbeatRequest{Lease: lease})
	elapsed := time.Since(started)
	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("row-lock heartbeat error = %v, want deadline exceeded", err)
	}
	minimum := schemaexecutor.DatabaseOperationTimeout - time.Second
	maximum := schemaexecutor.DatabaseOperationTimeout + 5*time.Second
	if elapsed < minimum || elapsed > maximum {
		t.Fatalf("row-lock timeout = %s, want between %s and %s", elapsed, minimum, maximum)
	}
	if err := transaction.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	expireSchemaExecutorJob(t, ctx, environment, jobID)
	recovered, err := environment.store.RecoverExpired(ctx)
	if err != nil || !recovered {
		t.Fatalf("recover ambiguous timed-out heartbeat: recovered=%t err=%v", recovered, err)
	}
	retryRecord, retryLease, _ := environment.startPlan(t, ctx,
		schemaExecutorIntegrationUUID(744), schemaExecutorIntegrationUUID(745))
	if retryRecord.Job.JobID != jobID || retryRecord.Lease.JobAttempt != 2 {
		t.Fatalf("timeout retry job=%s attempt=%d",
			retryRecord.Job.JobID, retryRecord.Lease.JobAttempt)
	}
	if err := environment.store.RecordFailure(ctx, schemaexecutor.FailureRequest{
		Lease: retryLease, State: schemaexecutor.FailureTerminal,
		Code:                 "integration_timeout_complete",
		RedactedDetailSHA256: schemaExecutorIntegrationDetailSHA256,
	}); err != nil {
		t.Fatalf("finish timeout test job: %v", err)
	}
}
