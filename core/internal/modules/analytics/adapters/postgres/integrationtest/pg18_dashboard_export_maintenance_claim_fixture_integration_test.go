//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

type analyticsMaintenanceClaimFixture struct {
	candidate         platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate
	cleanupGeneration int64
	deleteGeneration  int64
	cleanupOwner      uuid.UUID
	localSHA256       [32]byte
	localAt           time.Time
	resolutionID      uuid.UUID
	resolutionSHA256  [32]byte
	effectReceiptID   uuid.UUID
	effectReceiptSHA  [32]byte
	resolvedAt        time.Time
}

func claimAnalyticsMaintenanceUnknown(t *testing.T, ctx context.Context,
	fixture analyticsMaintenanceProtocolFixture,
) analyticsMaintenanceClaimFixture {
	t.Helper()
	connection := analyticsRuntimeConnection(t, ctx, fixture.database)
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	candidate := fixture.candidate
	cleanupOwner := uuid.MustParse(analyticsMaintenanceCleanup)
	var disposition, mode, terminal string
	var cleanupGeneration, deleteGeneration int64
	var leaseUntil, localAt time.Time
	var localDigest []byte
	err := transaction.QueryRowContext(ctx, `SELECT claim_state,cleanup_generation,
lease_until,cleanup_mode,terminal_state,delete_attempt_generation,
local_state_sha256,local_state_at
FROM analytics_dashboard_export_maintenance_unknown_claim_v1(
't_analytics_a',$1,$2,'t_analytics_a',$3,'analytics',1,1,$4,$5,1,1,$6,$7,$8,$9,$10,$11)`,
		candidate.Kind, candidate.TenantID, candidate.ExportID, candidate.JobID,
		candidate.AttemptID, candidate.OriginalLeaseOwnerID, candidate.EffectID,
		candidate.EffectRequestSHA256[:], candidate.UnknownAt, cleanupOwner,
		fixture.observedAt).Scan(&disposition, &cleanupGeneration, &leaseUntil,
		&mode, &terminal, &deleteGeneration, &localDigest, &localAt)
	if err != nil || disposition != "claimed" || mode != "completed" ||
		terminal != "deleted" || cleanupGeneration < 1 ||
		!leaseUntil.After(fixture.observedAt) || len(localDigest) != 32 {
		t.Fatalf("claim Analytics maintenance = %s/%s/%s gen=%d: %v",
			disposition, mode, terminal, cleanupGeneration, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	claim := analyticsMaintenanceClaimFixture{candidate: candidate,
		cleanupGeneration: cleanupGeneration, deleteGeneration: deleteGeneration,
		cleanupOwner: cleanupOwner, localAt: localAt.UTC(),
		resolutionID: uuid.MustParse(analyticsMaintenanceResolution),
		resolvedAt:   fixture.observedAt.Add(time.Second).UTC().Truncate(time.Microsecond)}
	copy(claim.localSHA256[:], localDigest)
	claim.effectReceiptID, err =
		platformasync.AnalyticsDashboardExportMaintenanceReceiptID(candidate)
	if err == nil {
		claim.effectReceiptSHA, err =
			platformasync.AnalyticsDashboardExportMaintenanceReceiptSHA256(candidate,
				claim.effectReceiptID, "deleted", claim.resolvedAt)
	}
	if err != nil {
		t.Fatal(err)
	}
	claim.resolutionSHA256 = analyticsMaintenanceResolutionSHA(t, ctx,
		fixture.database, claim)
	return claim
}

func analyticsMaintenanceResolutionSHA(t *testing.T, ctx context.Context,
	database *sql.DB, claim analyticsMaintenanceClaimFixture,
) (digest [32]byte) {
	t.Helper()
	var encoded []byte
	err := database.QueryRowContext(ctx, `SELECT
t_analytics_a.analytics_dashboard_export_maintenance_resolution_digest_v1(
'absent','deleted',NULL,NULL,NULL,NULL,NULL,NULL,NULL,$1,
NULL,NULL,NULL,$2,$3)`, claim.localSHA256[:], claim.localAt,
		claim.resolvedAt).Scan(&encoded)
	if err != nil || len(encoded) != 32 {
		t.Fatalf("compute Analytics maintenance resolution digest: %v", err)
	}
	copy(digest[:], encoded)
	return digest
}

const analyticsMaintenanceResolveSQL = `SELECT
analytics_dashboard_export_maintenance_unknown_resolve_v1(
't_analytics_a',$1,$2,'t_analytics_a',$3,'analytics',1,1,$4,$5,1,1,$6,$7,$8,$9,
$10,$11,$12,$13,$14,$15,'absent',$16,'deleted',$17,$18,
NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,$19)`

func analyticsMaintenanceResolveArguments(claim analyticsMaintenanceClaimFixture,
	effectReceiptID uuid.UUID, effectReceiptSHA, resolutionSHA [32]byte,
) []any {
	candidate := claim.candidate
	return []any{candidate.Kind, candidate.TenantID, candidate.ExportID,
		candidate.JobID, candidate.AttemptID, candidate.OriginalLeaseOwnerID,
		candidate.EffectID, candidate.EffectRequestSHA256[:], candidate.UnknownAt,
		claim.cleanupGeneration, claim.cleanupOwner, claim.deleteGeneration,
		claim.resolutionID, effectReceiptID, effectReceiptSHA[:], resolutionSHA[:],
		claim.localSHA256[:], claim.localAt, claim.resolvedAt}
}

func runAnalyticsMaintenanceResolve(ctx context.Context, connection *sql.Conn,
	claim analyticsMaintenanceClaimFixture, effectReceiptID uuid.UUID,
	effectReceiptSHA, resolutionSHA [32]byte,
) (string, error) {
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(ctx, `SET LOCAL search_path=t_analytics_a,platform,pg_catalog;
SET LOCAL lock_timeout='10s'; SET LOCAL statement_timeout='20s'`); err != nil {
		return "", err
	}
	var state string
	err = transaction.QueryRowContext(ctx, analyticsMaintenanceResolveSQL,
		analyticsMaintenanceResolveArguments(claim, effectReceiptID,
			effectReceiptSHA, resolutionSHA)...).Scan(&state)
	if err == nil {
		err = transaction.Commit()
	}
	return state, err
}

func analyticsMaintenanceSnapshot(t *testing.T, ctx context.Context,
	database *sql.DB, claim analyticsMaintenanceClaimFixture,
) string {
	t.Helper()
	var snapshot string
	err := database.QueryRowContext(ctx, `SELECT jsonb_build_object(
'tenant',(SELECT to_jsonb(item) FROM t_analytics_a.analytics_dashboard_exports item
 WHERE item.export_id=$1),'quota',(SELECT to_jsonb(quota)
 FROM platform.artifact_quota_counters quota WHERE quota.tenant_id=$2),
'job',(SELECT to_jsonb(job) FROM platform.jobs job WHERE job.id=$3),
'attempt',(SELECT to_jsonb(attempt) FROM platform.job_attempts attempt
 WHERE attempt.attempt_id=$4),'platform_receipts',(SELECT count(*)
 FROM platform.idempotency_receipts receipt WHERE receipt.causation_id=$3),
'platform_resolutions',(SELECT count(*) FROM
 platform.analytics_dashboard_export_maintenance_unknown_resolutions WHERE job_id=$3),
'tenant_resolutions',(SELECT count(*) FROM
 t_analytics_a.analytics_dashboard_export_maintenance_unknown_resolutions WHERE job_id=$3)
)::text`, claim.candidate.ExportID, claim.candidate.TenantID, claim.candidate.JobID,
		claim.candidate.AttemptID).Scan(&snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}
