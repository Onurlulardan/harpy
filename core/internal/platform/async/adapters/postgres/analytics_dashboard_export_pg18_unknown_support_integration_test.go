//go:build integration

package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

type analyticsDashboardExportUnknownEvidence struct {
	exportID, jobID, attemptID, leaseID, effectID uuid.UUID
	request                                       []byte
	unknownAt                                     time.Time
}

func createAnalyticsDashboardExportUnknown(
	t *testing.T, fixture analyticsDashboardExportPG18,
) analyticsDashboardExportUnknownEvidence {
	t.Helper()
	evidence := analyticsDashboardExportUnknownEvidence{
		exportID: analyticsExportV7(t), jobID: analyticsExportV7(t),
		leaseID: analyticsExportV7(t), effectID: analyticsExportV7(t),
		request: analyticsExportDigest(3),
	}
	seedAnalyticsDashboardExportJob(t, fixture.ctx, fixture.runtime,
		evidence.exportID, evidence.jobID, time.Now().UTC().Add(-time.Second))
	claim := claimAnalyticsDashboardExport(t, fixture.ctx, fixture.runtime,
		evidence.exportID, evidence.jobID, evidence.leaseID, 1)
	evidence.attemptID = analyticsExportAttemptID(t, claim)
	startAnalyticsDashboardExportEffect(t, fixture, evidence.exportID,
		evidence.jobID, evidence.attemptID, evidence.leaseID, evidence.effectID,
		evidence.request)
	setAnalyticsDashboardExportLease(
		t, fixture, evidence.jobID, time.Now().UTC().Add(-time.Second),
	)
	var state string
	var failure []byte
	err := fixture.runtime.QueryRow(fixture.ctx, `SELECT * FROM
        tenant_a.test_analytics_export_expired_finalize(
            $1,$2,$3,1,$4,$5,$6,$7)`, evidence.exportID, evidence.jobID,
		evidence.attemptID, evidence.leaseID, evidence.effectID, evidence.request,
		time.Now().UTC()).Scan(&state, &evidence.unknownAt, &failure)
	if err != nil || state != "outcome_unknown" || len(failure) != 0 {
		t.Fatalf("post-effect expiry state=%s digest=%d err=%v", state, len(failure), err)
	}
	return evidence
}
