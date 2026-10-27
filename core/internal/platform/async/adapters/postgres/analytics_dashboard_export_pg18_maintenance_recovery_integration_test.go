//go:build integration

package postgres

import (
	"bytes"
	"testing"
	"time"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

type analyticsMaintenanceResolution struct {
	receiptID, effectReceiptID uuid.UUID
	receiptSHA, resolutionSHA  [32]byte
	terminal                   string
	resolvedAt                 time.Time
}

func TestAnalyticsDashboardExportMaintenanceRecoveryOnPG184(t *testing.T) {
	fixture := startAnalyticsDashboardExportPG18(t)
	prepareAnalyticsMaintenancePG18(t, fixture)
	if _, err := fixture.database.ExecContext(fixture.ctx,
		analyticsDashboardExportMaintenanceRecoveryWrappersPG18); err != nil {
		t.Fatalf("prepare maintenance recovery wrappers: %v", err)
	}
	assertAnalyticsMaintenanceSHA1Vectors(t, fixture)

	_, resolveCandidate := createAnalyticsMaintenanceUnknownCandidate(t, fixture,
		platformasync.AnalyticsDashboardExportDeleteJobKind, false)
	resolveClaim := claimAnalyticsMaintenanceUnknown(t, fixture, resolveCandidate)
	resolveClaim.leaseUntil = checkpointAnalyticsMaintenanceUnknown(t, fixture,
		resolveCandidate, resolveClaim)

	_, deferCandidate := createAnalyticsMaintenanceUnknownCandidate(t, fixture,
		platformasync.AnalyticsDashboardExportExpireJobKind, true)
	deferClaim := claimAnalyticsMaintenanceUnknown(t, fixture, deferCandidate)
	deferAnalyticsMaintenanceUnknown(t, fixture, deferCandidate, deferClaim)
	expireParity := newAnalyticsMaintenanceResolution(t, deferCandidate, "expired")
	assertAnalyticsMaintenanceCanonicalParity(t, fixture, deferCandidate, expireParity)

	_, movedCandidate := createAnalyticsMaintenanceUnknownCandidate(t, fixture,
		platformasync.AnalyticsDashboardExportExpireJobKind, false)
	movedClaim := claimAnalyticsMaintenanceUnknown(t, fixture, movedCandidate)
	_, forgedCandidate := createAnalyticsMaintenanceUnknownCandidate(t, fixture,
		platformasync.AnalyticsDashboardExportDeleteJobKind, false)
	forgedClaim := claimAnalyticsMaintenanceUnknown(t, fixture, forgedCandidate)

	if _, err := fixture.database.ExecContext(fixture.ctx, `UPDATE platform.cutover_groups
SET generation=2 WHERE scope_kind='tenant' AND scope_id=$1 AND group_id='analytics'`,
		asyncTenantA); err != nil {
		t.Fatalf("bump current V2 authority generation: %v", err)
	}
	if _, err := fixture.database.ExecContext(fixture.ctx, `UPDATE platform.domain_cutovers
SET generation=2 WHERE scope_kind='tenant' AND scope_id=$1 AND domain='analytics'`,
		asyncTenantA); err != nil {
		t.Fatalf("bump current V2 domain generation: %v", err)
	}
	beforeReserved, beforeRevision := analyticsDashboardExportQuota(t, fixture)
	resolution := newAnalyticsMaintenanceResolution(t, resolveCandidate, "deleted")
	assertAnalyticsMaintenanceCanonicalParity(t, fixture, resolveCandidate, resolution)
	replayed, receiptID, state, err := callAnalyticsMaintenanceResolve(
		fixture, resolveCandidate, resolveClaim, resolution,
	)
	if err != nil || replayed || receiptID != resolution.effectReceiptID || state != "succeeded" {
		t.Fatalf("resolve maintenance replayed=%v receipt=%s state=%s err=%v",
			replayed, receiptID, state, err)
	}
	afterReserved, afterRevision := analyticsDashboardExportQuota(t, fixture)
	if beforeReserved != afterReserved || beforeRevision != afterRevision {
		t.Fatalf("Platform resolution mutated quota: before=(%d,%d) after=(%d,%d)",
			beforeReserved, beforeRevision, afterReserved, afterRevision)
	}
	assertAnalyticsMaintenanceResolutionRows(t, fixture, resolveCandidate, resolution)

	replayed, receiptID, state, err = callAnalyticsMaintenanceResolve(
		fixture, resolveCandidate, resolveClaim, resolution,
	)
	if err != nil || !replayed || receiptID != resolution.effectReceiptID || state != "succeeded" {
		t.Fatalf("exact resolution replay=%v receipt=%s state=%s err=%v",
			replayed, receiptID, state, err)
	}
	assertAnalyticsMaintenanceResolvedDisposition(t, fixture, resolveCandidate,
		resolveClaim.owner)
	before := analyticsDashboardExportSnapshot(t, fixture, resolveCandidate.JobID)
	tampered := resolution
	tampered.effectReceiptID = uuid.NewSHA1(uuid.NameSpaceOID, []byte("noncanonical-replay-receipt"))
	tampered.receiptSHA = analyticsMaintenanceSQLReceiptDigest(t, fixture, resolveCandidate,
		tampered.effectReceiptID, tampered.terminal, tampered.resolvedAt)
	_, _, _, err = callAnalyticsMaintenanceResolve(
		fixture, resolveCandidate, resolveClaim, tampered,
	)
	after := analyticsDashboardExportSnapshot(t, fixture, resolveCandidate.JobID)
	if analyticsExportSQLState(err) != "P6048" || before != after {
		t.Fatalf("noncanonical receipt replay state=%s changed=%v err=%v",
			analyticsExportSQLState(err), before != after, err)
	}
	forgedResolution := arbitraryAnalyticsMaintenanceResolution(
		t, fixture, forgedCandidate, "deleted",
	)
	before = analyticsDashboardExportSnapshot(t, fixture, forgedCandidate.JobID)
	_, _, _, err = callAnalyticsMaintenanceResolve(
		fixture, forgedCandidate, forgedClaim, forgedResolution,
	)
	after = analyticsDashboardExportSnapshot(t, fixture, forgedCandidate.JobID)
	if analyticsExportSQLState(err) != "P6048" || before != after {
		t.Fatalf("arbitrary UUIDv5 receipt state=%s changed=%v err=%v",
			analyticsExportSQLState(err), before != after, err)
	}

	if _, err = fixture.database.ExecContext(fixture.ctx, `UPDATE platform.cutover_groups
SET state='v1',writer='v1' WHERE scope_kind='tenant' AND scope_id=$1
AND group_id='analytics'`, asyncTenantA); err != nil {
		t.Fatalf("move maintenance writer: %v", err)
	}
	if _, err = fixture.database.ExecContext(fixture.ctx, `UPDATE platform.domain_cutovers
SET state='v1',writer='v1' WHERE scope_kind='tenant' AND scope_id=$1
AND domain='analytics'`, asyncTenantA); err != nil {
		t.Fatalf("move maintenance domain writer: %v", err)
	}
	movedResolution := newAnalyticsMaintenanceResolution(t, movedCandidate, "expired")
	before = analyticsDashboardExportSnapshot(t, fixture, movedCandidate.JobID)
	_, _, _, err = callAnalyticsMaintenanceResolve(
		fixture, movedCandidate, movedClaim, movedResolution,
	)
	after = analyticsDashboardExportSnapshot(t, fixture, movedCandidate.JobID)
	if analyticsExportSQLState(err) != "42501" || before != after {
		t.Fatalf("moved writer state=%s changed=%v err=%v",
			analyticsExportSQLState(err), before != after, err)
	}
}

func newAnalyticsMaintenanceResolution(t *testing.T,
	candidate platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate,
	terminal string,
) analyticsMaintenanceResolution {
	t.Helper()
	resolution := analyticsMaintenanceResolution{
		receiptID: analyticsExportV7(t), terminal: terminal,
		resolvedAt: time.Now().UTC().Truncate(time.Microsecond),
	}
	var err error
	resolution.effectReceiptID, err =
		platformasync.AnalyticsDashboardExportMaintenanceReceiptID(candidate)
	if err != nil {
		t.Fatal(err)
	}
	resolution.receiptSHA, err =
		platformasync.AnalyticsDashboardExportMaintenanceReceiptSHA256(
			candidate, resolution.effectReceiptID, terminal, resolution.resolvedAt)
	if err != nil {
		t.Fatal(err)
	}
	resolution.resolutionSHA[0] = 17
	return resolution
}

func callAnalyticsMaintenanceResolve(fixture analyticsDashboardExportPG18,
	candidate platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate,
	claim analyticsMaintenanceClaim, resolution analyticsMaintenanceResolution,
) (bool, uuid.UUID, string, error) {
	arguments := append(analyticsMaintenanceCandidateArguments(candidate),
		claim.generation, claim.owner, resolution.receiptID, resolution.effectReceiptID,
		resolution.receiptSHA[:], resolution.terminal, resolution.resolutionSHA[:],
		resolution.resolvedAt)
	var replayed bool
	var receiptID uuid.UUID
	var state string
	err := fixture.runtime.QueryRow(fixture.ctx, `SELECT * FROM
tenant_a.test_analytics_maintenance_unknown_resolve(
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,
    $19,$20,$21,$22)`, arguments...).Scan(&replayed, &receiptID, &state)
	return replayed, receiptID, state, err
}

func assertAnalyticsMaintenanceCanonicalParity(t *testing.T,
	fixture analyticsDashboardExportPG18,
	candidate platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate,
	resolution analyticsMaintenanceResolution,
) {
	t.Helper()
	var sqlID uuid.UUID
	var sqlDigest []byte
	err := fixture.database.QueryRowContext(fixture.ctx, `SELECT
platform.platform_async_analytics_export_maint_receipt_id_v1(
    attempt_row.effect_id,job.kind),
platform.platform_async_analytics_dashboard_export_maintenance_receipt_sha256_v1(
    job,attempt_row,$2,$3,$4)
FROM platform.jobs job JOIN platform.job_attempts attempt_row
  ON attempt_row.job_id=job.id
WHERE job.id=$1 AND attempt_row.attempt=job.attempts`, candidate.JobID,
		resolution.effectReceiptID, resolution.terminal, resolution.resolvedAt).Scan(
		&sqlID, &sqlDigest)
	if err != nil || sqlID != resolution.effectReceiptID ||
		!bytes.Equal(sqlDigest, resolution.receiptSHA[:]) {
		t.Fatalf("Go/SQL receipt parity id=%s digest=%x err=%v",
			sqlID, sqlDigest, err)
	}
}
