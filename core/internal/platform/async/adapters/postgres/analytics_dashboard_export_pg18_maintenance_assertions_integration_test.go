//go:build integration

package postgres

import (
	"testing"
	"time"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func assertAnalyticsMaintenanceResolutionRows(t *testing.T,
	fixture analyticsDashboardExportPG18,
	candidate platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate,
	resolution analyticsMaintenanceResolution,
) {
	t.Helper()
	var exact bool
	err := fixture.database.QueryRowContext(fixture.ctx, `SELECT
    job.state='succeeded' AND job.completed_at=$3 AND job.updated_at=$3
    AND job.last_error_code IS NULL AND job.retain_until=$3+interval '90 days'
    AND attempt.outcome='succeeded' AND attempt.finished_at=$3
    AND attempt.error_code IS NULL AND attempt.effect_receipt_id=$2
    AND marker.kind=$4 AND marker.original_unknown_at=$5
    AND marker.original_attempt_outcome='dead'
    AND marker.original_error_code='JOB_OUTCOME_UNKNOWN'
    AND marker.effect_receipt_id=$2 AND marker.effect_receipt_sha256=$6
    AND marker.terminal_state=$7 AND marker.resolved_at=$3
    AND receipt.receipt_id=$2 AND receipt.effect_id=$8
    AND receipt.effect_status='succeeded' AND receipt.completed_at=$3
FROM platform.jobs job JOIN platform.job_attempts attempt ON attempt.job_id=job.id
JOIN platform.analytics_dashboard_export_maintenance_unknown_resolutions marker
  ON marker.job_id=job.id
JOIN platform.idempotency_receipts receipt ON receipt.receipt_id=marker.effect_receipt_id
WHERE job.id=$1 AND attempt.attempt=job.attempts`, candidate.JobID,
		resolution.effectReceiptID, resolution.resolvedAt, candidate.Kind,
		candidate.UnknownAt, fmtDigest(resolution.receiptSHA), resolution.terminal,
		candidate.EffectID).Scan(&exact)
	if err != nil || !exact {
		t.Fatalf("maintenance resolution rows exact=%v err=%v", exact, err)
	}
}

func assertAnalyticsMaintenanceResolvedDisposition(t *testing.T,
	fixture analyticsDashboardExportPG18,
	candidate platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate,
	cleanupOwner uuid.UUID,
) {
	t.Helper()
	arguments := append(analyticsMaintenanceCandidateArguments(candidate),
		cleanupOwner, time.Now().UTC())
	var state string
	var generation pgtype.Int8
	var leaseUntil pgtype.Timestamptz
	err := fixture.runtime.QueryRow(fixture.ctx, `SELECT * FROM
tenant_a.test_analytics_maintenance_unknown_claim(
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		arguments...).Scan(&state, &generation, &leaseUntil)
	if err != nil || state != "resolved" || generation.Valid || leaseUntil.Valid {
		t.Fatalf("resolved disposition=%s generation=%+v lease=%+v err=%v",
			state, generation, leaseUntil, err)
	}
}

func arbitraryAnalyticsMaintenanceResolution(t *testing.T,
	fixture analyticsDashboardExportPG18,
	candidate platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate,
	terminal string,
) analyticsMaintenanceResolution {
	t.Helper()
	resolution := analyticsMaintenanceResolution{
		receiptID: analyticsExportV7(t),
		effectReceiptID: uuid.NewSHA1(uuid.NameSpaceOID,
			[]byte("arbitrary-maintenance-receipt")),
		terminal: terminal, resolvedAt: time.Now().UTC().Truncate(time.Microsecond),
	}
	resolution.resolutionSHA[0] = 23
	resolution.receiptSHA = analyticsMaintenanceSQLReceiptDigest(t, fixture, candidate,
		resolution.effectReceiptID, terminal, resolution.resolvedAt)
	return resolution
}

func analyticsMaintenanceSQLReceiptDigest(t *testing.T,
	fixture analyticsDashboardExportPG18,
	candidate platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate,
	receiptID uuid.UUID, terminal string, resolvedAt time.Time,
) [32]byte {
	t.Helper()
	var digest []byte
	err := fixture.database.QueryRowContext(fixture.ctx, `SELECT
platform.platform_async_analytics_dashboard_export_maintenance_receipt_sha256_v1(
    job,attempt_row,$2,$3,$4)
FROM platform.jobs job JOIN platform.job_attempts attempt_row
  ON attempt_row.job_id=job.id
WHERE job.id=$1 AND attempt_row.attempt=job.attempts`, candidate.JobID,
		receiptID, terminal, resolvedAt).Scan(&digest)
	if err != nil || len(digest) != 32 {
		t.Fatalf("recompute arbitrary receipt digest length=%d err=%v", len(digest), err)
	}
	var result [32]byte
	copy(result[:], digest)
	return result
}

func fmtDigest(value [32]byte) string {
	const hex = "0123456789abcdef"
	result := make([]byte, 64)
	for index, valueByte := range value {
		result[index*2], result[index*2+1] = hex[valueByte>>4], hex[valueByte&15]
	}
	return string(result)
}

func assertAnalyticsMaintenanceSHA1Vectors(
	t *testing.T, fixture analyticsDashboardExportPG18,
) {
	t.Helper()
	var empty, abc, twoBlock string
	err := fixture.database.QueryRowContext(fixture.ctx, `SELECT
encode(platform.platform_async_analytics_export_sha1_v1(''::bytea),'hex'),
encode(platform.platform_async_analytics_export_sha1_v1(convert_to('abc','UTF8')),'hex'),
encode(platform.platform_async_analytics_export_sha1_v1(
    convert_to(repeat('a',56),'UTF8')),'hex')`).Scan(&empty, &abc, &twoBlock)
	if err != nil || empty != "da39a3ee5e6b4b0d3255bfef95601890afd80709" ||
		abc != "a9993e364706816aba3e25717850c26c9cd0d89d" ||
		twoBlock != "c2db330f6083854c99d4b5bfb6e8f29f201be699" {
		t.Fatalf("pure SQL SHA-1 vectors empty=%s abc=%s two=%s err=%v",
			empty, abc, twoBlock, err)
	}
}
