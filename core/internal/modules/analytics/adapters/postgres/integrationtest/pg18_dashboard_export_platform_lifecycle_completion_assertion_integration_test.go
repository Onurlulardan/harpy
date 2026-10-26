//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
)

func assertAnalyticsDashboardExportMaintenanceTerminal(t *testing.T,
	ctx context.Context, database *sql.DB, jobID, attemptID, receiptID string,
	request effectsv1.Request, code, resultReference, requestSHA string,
	completedAt time.Time,
) {
	t.Helper()
	var jobState, attemptOutcome, storedReceipt, provider, operation string
	var storedCode, storedResult, storedRequest, effectKind, effectID string
	var leaseCleared, jobClockExact, attemptClockExact, receiptExact bool
	err := database.QueryRowContext(ctx, `SELECT job.state,
job.lease_owner_id IS NULL AND job.lease_heartbeat_at IS NULL AND job.lease_until IS NULL,
job.completed_at=job.updated_at AND job.completed_at=$4 AND
 job.retain_until=$4::timestamptz+interval '90 days',
attempt.outcome,attempt.finished_at=job.completed_at,attempt.effect_receipt_id::text,
receipt.provider,receipt.operation,receipt.effect_code,receipt.effect_result_ref,
receipt.request_sha256,receipt.effect_kind,receipt.effect_id::text,
receipt.receipt_id::text=$3 AND receipt.effect_status='succeeded' AND
 receipt.effect_error_class IS NULL AND receipt.effect_result_sha256=encode(
 sha256(convert_to($5,'UTF8')),'hex') AND receipt.provider_reference=$6 AND
 receipt.correlation_id=job.id::text AND receipt.causation_id=job.id
FROM platform.jobs job JOIN platform.job_attempts attempt ON attempt.job_id=job.id
JOIN platform.idempotency_receipts receipt
 ON receipt.receipt_id=attempt.effect_receipt_id
WHERE job.id=$1::uuid AND attempt.attempt_id=$2::uuid`, jobID, attemptID, receiptID,
		completedAt, resultReference, request.PayloadReference).Scan(&jobState,
		&leaseCleared, &jobClockExact, &attemptOutcome, &attemptClockExact,
		&storedReceipt, &provider, &operation, &storedCode, &storedResult,
		&storedRequest, &effectKind, &effectID, &receiptExact)
	if err != nil || jobState != "succeeded" || !leaseCleared || !jobClockExact ||
		attemptOutcome != "succeeded" || !attemptClockExact || storedReceipt != receiptID ||
		provider != request.Provider || operation != request.Operation || storedCode != code ||
		storedResult != resultReference || storedRequest != requestSHA ||
		effectKind != request.EffectKind || effectID != request.EffectID || !receiptExact {
		t.Fatalf("maintenance terminal job=%s lease=%v clock=%v attempt=%s/%v "+
			"receipt=%s provider=%s/%s code=%s result=%s request=%s kind=%s effect=%s "+
			"exact=%v: %v", jobState, leaseCleared, jobClockExact, attemptOutcome,
			attemptClockExact, storedReceipt, provider, operation, storedCode,
			storedResult, storedRequest, effectKind, effectID, receiptExact, err)
	}
}

func analyticsDashboardExportPlatformRootSnapshot(t *testing.T,
	ctx context.Context, database *sql.DB, jobID string,
) string {
	t.Helper()
	var snapshot string
	err := database.QueryRowContext(ctx, `SELECT jsonb_build_object(
'job',(SELECT to_jsonb(job) FROM platform.jobs job WHERE job.id=$1),
'attempts',(SELECT COALESCE(jsonb_agg(to_jsonb(attempt) ORDER BY attempt.attempt),
 '[]'::jsonb) FROM platform.job_attempts attempt WHERE attempt.job_id=$1),
'receipts',(SELECT COALESCE(jsonb_agg(to_jsonb(receipt) ORDER BY receipt.receipt_id),
 '[]'::jsonb) FROM platform.idempotency_receipts receipt
 WHERE receipt.causation_id=$1))::text`, jobID).Scan(&snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}
