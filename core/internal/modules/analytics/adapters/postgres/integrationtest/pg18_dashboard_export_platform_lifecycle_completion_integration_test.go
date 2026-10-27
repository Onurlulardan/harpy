//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

func completeAnalyticsDashboardExportPlatformMaintenance(t *testing.T,
	ctx context.Context, database *sql.DB, connection *sql.Conn, exportID,
	jobID, attemptID, leaseOwner, kind, terminalState string,
) string {
	t.Helper()
	var effectStarted time.Time
	var requestSHA string
	err := database.QueryRowContext(ctx, `SELECT job.effect_started_at,
attempt.effect_request_sha256 FROM platform.jobs job
JOIN platform.job_attempts attempt ON attempt.job_id=job.id
WHERE job.id=$1 AND attempt.attempt_id=$2`, jobID, attemptID).Scan(
		&effectStarted, &requestSHA)
	if err != nil {
		t.Fatal(err)
	}
	request, canonicalRequestSHA := analyticsDashboardExportMaintenanceRequest(t,
		exportID, jobID, attemptID, kind, effectStarted)
	if requestSHA != canonicalRequestSHA {
		t.Fatalf("maintenance request SHA = %s, want %s", requestSHA,
			canonicalRequestSHA)
	}
	completedAt := time.Now().UTC().Add(100 * time.Millisecond).Truncate(time.Microsecond)
	if !completedAt.After(effectStarted) {
		completedAt = effectStarted.Add(time.Microsecond)
	}
	candidate := analyticsDashboardExportMaintenanceCandidate(t, request,
		exportID, jobID, attemptID, leaseOwner, kind, effectStarted, requestSHA)
	receiptID, err := platformasync.AnalyticsDashboardExportMaintenanceReceiptID(candidate)
	if err != nil {
		t.Fatal(err)
	}
	receiptSHA, err := platformasync.AnalyticsDashboardExportMaintenanceReceiptSHA256(
		candidate, receiptID, terminalState, completedAt)
	if err != nil {
		t.Fatal(err)
	}
	operation, _ := analyticsDashboardExportMaintenanceOperation(t, kind)
	code := "ANALYTICS_EXPORT_DELETED"
	if operation == "expire" {
		code = "ANALYTICS_EXPORT_EXPIRY_COMPLETE"
	}
	providerReference := exportID
	resultReference := "analytics-dashboard-export://" + operation + "/" +
		exportID + "/" + terminalState
	resultDigest := sha256.Sum256([]byte(resultReference))
	receipt, err := effectsv1.NewSucceededReceipt(request, receiptID.String(), code,
		&providerReference, resultReference, hex.EncodeToString(resultDigest[:]),
		completedAt.Format(time.RFC3339Nano))
	if err != nil || receipt.ValidateForRequest(request) != nil {
		t.Fatalf("canonical maintenance receipt: %v", err)
	}
	storeSQL := loadAnalyticsDashboardExportProductionSource(t,
		"platform/async/adapters/postgres/sql/store_external_effect_receipt.sql")
	lockSQL := loadAnalyticsDashboardExportProductionSource(t,
		"platform/async/adapters/postgres/sql/lock_job_claim.sql")
	attemptSQL := loadAnalyticsDashboardExportProductionSource(t,
		"platform/async/adapters/postgres/sql/complete_external_job_attempt.sql")
	jobSQL := loadAnalyticsDashboardExportProductionSource(t,
		"platform/async/adapters/postgres/sql/complete_external_job.sql")
	beforeCompletion := analyticsDashboardExportLifecycleSnapshot(t, ctx, database,
		exportID, jobID)
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback() }()
	var cancelRequested sql.NullTime
	var lockedEffectStarted, leaseUntil time.Time
	err = transaction.QueryRowContext(ctx, lockSQL, "tenant", analyticsTenantA,
		"analytics", int64(1), "analytics", int64(1), jobID, leaseOwner,
		int64(1)).Scan(&cancelRequested, &lockedEffectStarted, &leaseUntil)
	if err != nil || cancelRequested.Valid || !lockedEffectStarted.Equal(effectStarted) ||
		!leaseUntil.After(time.Now().UTC()) {
		t.Fatalf("lock production maintenance claim = %v/%v/%v: %v",
			cancelRequested, lockedEffectStarted, leaseUntil, err)
	}
	var exact bool
	err = transaction.QueryRowContext(ctx, storeSQL, receiptID, analyticsTenantA,
		"tenant", analyticsTenantA, "analytics", int64(1), "analytics", int64(1),
		requestSHA, request.EffectID, kind, 1, kind+":"+exportID,
		"analytics.dashboard-export", operation, jobID, jobID, "succeeded", nil,
		code, providerReference, resultReference, hex.EncodeToString(resultDigest[:]),
		completedAt).Scan(&exact)
	if err != nil || !exact {
		_ = transaction.Rollback()
		if beforeCompletion != analyticsDashboardExportLifecycleSnapshot(t, ctx,
			database, exportID, jobID) {
			t.Fatal("failed maintenance receipt insert mutated lifecycle evidence")
		}
		t.Fatalf("store canonical maintenance receipt = %v: %v", exact, err)
	}
	var persistedAt time.Time
	err = transaction.QueryRowContext(ctx, attemptSQL, attemptID, jobID, leaseOwner,
		int64(1), effectStarted, receiptID, completedAt, request.EffectID,
		requestSHA).Scan(&persistedAt)
	if err != nil {
		t.Fatalf("complete maintenance attempt: %v", err)
	}
	var jobCompleted time.Time
	err = transaction.QueryRowContext(ctx, jobSQL, "tenant", analyticsTenantA,
		"analytics", int64(1), "analytics", int64(1), jobID, leaseOwner, int64(1),
		effectStarted, persistedAt).Scan(&jobCompleted)
	if err != nil || !jobCompleted.Equal(persistedAt) {
		t.Fatalf("complete maintenance job = %v/%v: %v", jobCompleted, persistedAt, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	assertAnalyticsDashboardExportMaintenanceTerminal(t, ctx, database, jobID,
		attemptID, receiptID.String(), request, code, resultReference, requestSHA,
		persistedAt)
	var liveSHA []byte
	err = database.QueryRowContext(ctx, `SELECT
platform.platform_async_analytics_dashboard_export_maintenance_receipt_sha256_v1(
job,attempt_row,$3,$4,$5) FROM platform.jobs job
JOIN platform.job_attempts attempt_row ON attempt_row.job_id=job.id
WHERE job.id=$1 AND attempt_row.attempt_id=$2`, jobID,
		attemptID, receiptID, terminalState, completedAt).Scan(&liveSHA)
	if err != nil || !bytes.Equal(liveSHA, receiptSHA[:]) {
		t.Fatalf("maintenance receipt SHA parity = %x/%x: %v", liveSHA, receiptSHA, err)
	}
	assertAnalyticsDashboardExportMaintenanceReceiptReplay(t, ctx, database,
		connection, exportID, jobID, analyticsDashboardExportMaintenanceReceipt{
			ID: receiptID, Request: request, Kind: kind, Operation: operation,
			Code: code, ProviderReference: providerReference,
			ResultReference: resultReference,
			ResultSHA:       hex.EncodeToString(resultDigest[:]),
			RequestSHA:      requestSHA, CompletedAt: completedAt,
		})
	return analyticsDashboardExportPlatformRootSnapshot(t, ctx, database, jobID)
}

func loadAnalyticsDashboardExportProductionSource(t *testing.T,
	relative string,
) string {
	t.Helper()
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate Analytics integration source")
	}
	internalRoot := filepath.Clean(filepath.Join(filepath.Dir(current),
		"..", "..", "..", "..", ".."))
	document, err := os.ReadFile(filepath.Join(internalRoot, filepath.FromSlash(relative)))
	if err != nil {
		t.Fatal(err)
	}
	return string(document)
}

func analyticsDashboardExportMaintenanceCandidate(t *testing.T,
	request effectsv1.Request, exportID, jobID, attemptID, leaseOwner, kind string,
	effectStarted time.Time, requestSHA string,
) platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate {
	t.Helper()
	digest, err := hex.DecodeString(requestSHA)
	if err != nil || len(digest) != 32 {
		t.Fatal("decode maintenance request SHA")
	}
	candidate := platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate{
		Kind: kind, TenantID: uuid.MustParse(analyticsTenantA), TenantSchema: "t_analytics_a",
		ExportID: uuid.MustParse(exportID), AuthorityGroupID: "analytics",
		AuthorityGroupGeneration: 1, CutoverGeneration: 1, JobID: uuid.MustParse(jobID),
		AttemptID: uuid.MustParse(attemptID), AttemptNumber: 1, LeaseGeneration: 1,
		OriginalLeaseOwnerID: uuid.MustParse(leaseOwner), EffectID: uuid.MustParse(request.EffectID),
		UnknownAt: effectStarted,
	}
	copy(candidate.EffectRequestSHA256[:], digest)
	return candidate
}
