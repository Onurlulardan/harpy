//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var analyticsMaintenanceReceiptNamespace = uuid.MustParse("8e7a6d02-df6d-4327-b179-160296b50b8e")

func analyticsMaintenanceSucceededReceipt(t *testing.T,
	evidence analyticsMaintenanceActive, state string,
) effectsv1.Receipt {
	t.Helper()
	request := evidence.permit.Request()
	code := "ANALYTICS_EXPORT_DELETED"
	if evidence.kind == "analytics.dashboard.export.expire" {
		code = "ANALYTICS_EXPORT_EXPIRY_COMPLETE"
	}
	resultRef := "analytics-dashboard-export://" + request.Operation + "/" +
		evidence.exportID.String() + "/" + state
	digest := sha256.Sum256([]byte(resultRef))
	providerRef := evidence.exportID.String()
	receipt, err := effectsv1.NewSucceededReceipt(request,
		analyticsMaintenanceCanonicalReceiptID(request, code), code, &providerRef,
		resultRef, hex.EncodeToString(digest[:]), time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		t.Fatal(err)
	}
	return receipt
}

func analyticsMaintenanceFailedReceipt(t *testing.T,
	evidence analyticsMaintenanceActive, code string, class effectsv1.ErrorClass,
) effectsv1.Receipt {
	t.Helper()
	request := evidence.permit.Request()
	providerRef := evidence.exportID.String()
	receipt, err := effectsv1.NewFailedReceipt(request,
		analyticsMaintenanceCanonicalReceiptID(request, code), code, class,
		&providerRef, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		t.Fatal(err)
	}
	return receipt
}

func analyticsMaintenanceCanonicalReceiptID(request effectsv1.Request, code string) string {
	return uuid.NewSHA1(analyticsMaintenanceReceiptNamespace,
		[]byte(request.EffectID+"\x00"+code)).String()
}

func storeAnalyticsMaintenanceReceipt(t *testing.T,
	fixture analyticsDashboardExportPG18, evidence analyticsMaintenanceActive,
	receipt effectsv1.Receipt,
) error {
	t.Helper()
	values, err := validateExternalReceipt(evidence.permit, receipt)
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := fixture.runtime.BeginTx(fixture.ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback(fixture.ctx)
	return storeExternalEffectReceipt(
		fixture.ctx, transaction, evidence.permit, receipt, values,
	)
}

func persistAnalyticsMaintenanceReceipt(ctx context.Context, database workerDatabase,
	permit ExternalEffectPermit, receipt effectsv1.Receipt, values externalReceiptValues,
) error {
	transaction, err := database.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer transaction.Rollback(ctx)
	if err = storeExternalEffectReceipt(ctx, transaction, permit, receipt, values); err != nil {
		return err
	}
	return transaction.Commit(ctx)
}

func analyticsMaintenanceReceiptSnapshot(t *testing.T,
	fixture analyticsDashboardExportPG18, jobID uuid.UUID,
) string {
	t.Helper()
	var receipts int
	if err := fixture.database.QueryRowContext(fixture.ctx, `SELECT count(*)
FROM platform.idempotency_receipts WHERE causation_id=$1`, jobID).Scan(&receipts); err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%s|receipts=%d",
		analyticsMaintenanceJobSnapshot(t, fixture.ctx, fixture, jobID), receipts)
}

func assertAnalyticsMaintenanceReceiptRows(t *testing.T,
	fixture analyticsDashboardExportPG18, evidence analyticsMaintenanceActive,
	receipt effectsv1.Receipt, jobState, attemptOutcome string,
) {
	t.Helper()
	var exact bool
	err := fixture.database.QueryRowContext(fixture.ctx, `SELECT
    job.state=$3 AND job.lease_owner_id IS NULL AND job.lease_until IS NULL
    AND attempt.outcome=$4 AND attempt.effect_receipt_id=$2
    AND receipt.receipt_id=$2 AND receipt.effect_id=attempt.effect_id
    AND receipt.request_sha256=attempt.effect_request_sha256
    AND receipt.effect_kind=job.kind AND receipt.effect_key=job.effect_idempotency_key
    AND receipt.provider='analytics.dashboard-export'
    AND receipt.operation=$5 AND receipt.correlation_id=job.id::text
    AND receipt.causation_id=job.id AND receipt.provider_reference=job.payload_ref
    AND receipt.effect_status=$6 AND receipt.effect_code=$7
FROM platform.jobs job JOIN platform.job_attempts attempt ON attempt.job_id=job.id
JOIN platform.idempotency_receipts receipt ON receipt.receipt_id=attempt.effect_receipt_id
WHERE job.id=$1 AND attempt.attempt=job.attempts`, evidence.jobID,
		receipt.ReceiptID, jobState, attemptOutcome, receipt.Operation,
		receipt.Status, receipt.Code).Scan(&exact)
	if err != nil || !exact {
		t.Fatalf("maintenance receipt rows exact=%v err=%v", exact, err)
	}
}
