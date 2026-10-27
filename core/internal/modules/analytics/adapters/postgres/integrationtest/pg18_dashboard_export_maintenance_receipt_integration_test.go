//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func TestAnalyticsDashboardExportMaintenanceReceiptParityAndTamperPG18(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	fixture := prepareAnalyticsMaintenanceProtocolPG18(t, ctx)
	claim := claimAnalyticsMaintenanceUnknown(t, ctx, fixture)
	var sqlDigest []byte
	err := fixture.database.QueryRowContext(ctx, `SELECT
platform.platform_async_analytics_dashboard_export_maintenance_receipt_sha256_v1(
job,attempt_row,$2,'deleted',$3)
FROM platform.jobs job JOIN platform.job_attempts attempt_row
  ON attempt_row.job_id=job.id
WHERE job.id=$1 AND attempt_row.attempt_id=$4`, claim.candidate.JobID,
		claim.effectReceiptID, claim.resolvedAt, claim.candidate.AttemptID).Scan(&sqlDigest)
	if err != nil || !bytes.Equal(sqlDigest, claim.effectReceiptSHA[:]) {
		t.Fatalf("Go/SQL maintenance receipt digest differs: sql=%x go=%x err=%v",
			sqlDigest, claim.effectReceiptSHA, err)
	}
	baseline := analyticsMaintenanceSnapshot(t, ctx, fixture.database, claim)
	connection := analyticsRuntimeConnection(t, ctx, fixture.database)
	tamperedID := analyticsMaintenanceEffectID("tampered-receipt")
	_, err = runAnalyticsMaintenanceResolve(ctx, connection, claim, tamperedID,
		claim.effectReceiptSHA, claim.resolutionSHA256)
	requireAnalyticsPGCode(t, err, "P6048", "tampered maintenance receipt identity")
	if after := analyticsMaintenanceSnapshot(t, ctx, fixture.database, claim); after != baseline {
		t.Fatal("tampered maintenance receipt identity mutated job, receipt, quota, or tenant")
	}
	tamperedDigest := claim.effectReceiptSHA
	tamperedDigest[0]++
	_, err = runAnalyticsMaintenanceResolve(ctx, connection, claim, claim.effectReceiptID,
		tamperedDigest, claim.resolutionSHA256)
	requireAnalyticsPGCode(t, err, "P6048", "tampered maintenance receipt digest")
	if after := analyticsMaintenanceSnapshot(t, ctx, fixture.database, claim); after != baseline {
		t.Fatal("tampered maintenance receipt digest mutated job, receipt, quota, or tenant")
	}
}
