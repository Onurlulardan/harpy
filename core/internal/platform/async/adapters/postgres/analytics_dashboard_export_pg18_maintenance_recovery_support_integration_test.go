//go:build integration

package postgres

import (
	"testing"
	"time"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

type analyticsMaintenanceClaim struct {
	state      string
	generation int64
	leaseUntil time.Time
	owner      uuid.UUID
}

func createAnalyticsMaintenanceUnknownCandidate(
	t *testing.T, fixture analyticsDashboardExportPG18, kind string, recoverExpired bool,
) (analyticsMaintenanceActive, platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate) {
	t.Helper()
	active := createAnalyticsMaintenanceActive(t, fixture, kind)
	if recoverExpired {
		setAnalyticsDashboardExportLease(t, fixture, active.jobID,
			time.Now().UTC().Add(-time.Second))
		result, err := active.repository.RecoverExpiredJobs(
			fixture.ctx, active.authority, 10,
		)
		if err != nil || result.Dead() != 1 || result.Total() != 1 {
			t.Fatalf("recover maintenance expiry result=%#v err=%v", result, err)
		}
	} else if _, err := active.repository.MarkExternalJobOutcomeUnknown(
		fixture.ctx, active.permit,
	); err != nil {
		t.Fatalf("mark maintenance unknown: %v", err)
	}
	candidate := nextAnalyticsMaintenanceUnknown(t, fixture)
	if candidate.JobID != active.jobID || candidate.AttemptID != active.attemptID ||
		candidate.EffectID != active.effectID || candidate.Kind != kind {
		t.Fatalf("maintenance candidate mismatch: %+v", candidate)
	}
	return active, candidate
}

func nextAnalyticsMaintenanceUnknown(
	t *testing.T, fixture analyticsDashboardExportPG18,
) platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate {
	t.Helper()
	candidate := platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate{}
	var request []byte
	err := fixture.runtime.QueryRow(fixture.ctx, `SELECT * FROM
platform.platform_async_analytics_dashboard_export_maintenance_unknown_candidate_v1()`).Scan(
		&candidate.Kind, &candidate.TenantID, &candidate.TenantSchema,
		&candidate.ExportID, &candidate.AuthorityGroupID,
		&candidate.AuthorityGroupGeneration, &candidate.CutoverGeneration,
		&candidate.JobID, &candidate.AttemptID, &candidate.AttemptNumber,
		&candidate.LeaseGeneration, &candidate.OriginalLeaseOwnerID,
		&candidate.EffectID, &request, &candidate.UnknownAt,
	)
	if err != nil || len(request) != 32 {
		t.Fatalf("read maintenance candidate digest=%d err=%v", len(request), err)
	}
	copy(candidate.EffectRequestSHA256[:], request)
	if !candidate.Valid() {
		t.Fatalf("invalid typed maintenance candidate: %+v", candidate)
	}
	return candidate
}

func analyticsMaintenanceCandidateArguments(
	candidate platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate,
) []any {
	return []any{
		candidate.Kind, candidate.TenantSchema, candidate.ExportID,
		candidate.AuthorityGroupID, candidate.AuthorityGroupGeneration,
		candidate.CutoverGeneration, candidate.JobID, candidate.AttemptID,
		candidate.AttemptNumber, candidate.LeaseGeneration,
		candidate.OriginalLeaseOwnerID, candidate.EffectID,
		candidate.EffectRequestSHA256[:], candidate.UnknownAt,
	}
}

func claimAnalyticsMaintenanceUnknown(
	t *testing.T, fixture analyticsDashboardExportPG18,
	candidate platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate,
) analyticsMaintenanceClaim {
	t.Helper()
	claim := analyticsMaintenanceClaim{owner: analyticsExportV7(t)}
	arguments := append(analyticsMaintenanceCandidateArguments(candidate),
		claim.owner, time.Now().UTC())
	err := fixture.runtime.QueryRow(fixture.ctx, `SELECT * FROM
tenant_a.test_analytics_maintenance_unknown_claim(
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		arguments...).Scan(&claim.state, &claim.generation, &claim.leaseUntil)
	if err != nil || claim.state != "claimed" || claim.generation < 1 ||
		!claim.leaseUntil.After(time.Now().UTC()) {
		t.Fatalf("claim maintenance unknown: %+v err=%v", claim, err)
	}
	return claim
}

func checkpointAnalyticsMaintenanceUnknown(
	t *testing.T, fixture analyticsDashboardExportPG18,
	candidate platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate,
	claim analyticsMaintenanceClaim,
) time.Time {
	t.Helper()
	arguments := append(analyticsMaintenanceCandidateArguments(candidate),
		claim.generation, claim.owner, time.Now().UTC())
	var leaseUntil time.Time
	err := fixture.runtime.QueryRow(fixture.ctx, `SELECT
tenant_a.test_analytics_maintenance_unknown_checkpoint(
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		arguments...).Scan(&leaseUntil)
	if err != nil || !leaseUntil.After(claim.leaseUntil) {
		t.Fatalf("checkpoint maintenance lease=%v previous=%v err=%v",
			leaseUntil, claim.leaseUntil, err)
	}
	return leaseUntil
}

func deferAnalyticsMaintenanceUnknown(
	t *testing.T, fixture analyticsDashboardExportPG18,
	candidate platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate,
	claim analyticsMaintenanceClaim,
) time.Time {
	t.Helper()
	arguments := append(analyticsMaintenanceCandidateArguments(candidate),
		claim.generation, claim.owner, time.Now().UTC())
	var retryAt time.Time
	err := fixture.runtime.QueryRow(fixture.ctx, `SELECT
tenant_a.test_analytics_maintenance_unknown_defer(
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		arguments...).Scan(&retryAt)
	if err != nil || !retryAt.After(time.Now().UTC().Add(25*time.Second)) {
		t.Fatalf("defer maintenance retry=%v err=%v", retryAt, err)
	}
	return retryAt
}
