//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"path"
	"sort"
	"strings"
	"testing"
	"time"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformmigration "github.com/Onurlulardan/harpy/core/internal/platform/migration"
	"github.com/google/uuid"
)

const (
	analyticsMaintenanceJob           = "019b2000-0000-7000-8000-000000000081"
	analyticsMaintenanceAttempt       = "019b2000-0000-7000-8000-000000000082"
	analyticsMaintenanceOwner         = "019b2000-0000-7000-8000-000000000083"
	analyticsMaintenanceCleanup       = "019b2000-0000-7000-8000-000000000084"
	analyticsMaintenanceResolution    = "019b2000-0000-7000-8000-000000000085"
	analyticsMaintenanceDeadJob       = "019b2000-0000-7000-8000-000000000091"
	analyticsMaintenanceDeadExport    = "019b2000-0000-7000-8000-000000000092"
	analyticsMaintenanceSuccessJob    = "019b2000-0000-7000-8000-000000000093"
	analyticsMaintenanceSuccessExport = "019b2000-0000-7000-8000-000000000094"
	analyticsMaintenanceRunningJob    = "019b2000-0000-7000-8000-000000000095"
	analyticsMaintenanceRunningExport = "019b2000-0000-7000-8000-000000000096"
)

type analyticsMaintenanceProtocolFixture struct {
	database   *sql.DB
	candidate  platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate
	observedAt time.Time
}

func prepareAnalyticsMaintenanceProtocolPG18(t *testing.T,
	ctx context.Context,
) analyticsMaintenanceProtocolFixture {
	t.Helper()
	database := prepareAnalyticsPlatformProtocolPG18(t, ctx)
	applyAnalyticsDocuments(t, ctx, database, "",
		analyticsMaintenancePlatformDocuments(t, "20260802020000", "20260802020024"))
	var exportUpdated time.Time
	if err := database.QueryRowContext(ctx, `SELECT updated_at FROM
t_analytics_a.analytics_dashboard_exports WHERE export_id=$1`,
		analyticsLockExport).Scan(&exportUpdated); err != nil {
		t.Fatal(err)
	}
	unknownAt := time.Now().UTC().Add(-100 * time.Millisecond).Truncate(time.Microsecond)
	if !unknownAt.After(exportUpdated.Add(3 * time.Microsecond)) {
		unknownAt = exportUpdated.Add(4 * time.Microsecond)
	}
	insertAnalyticsMaintenanceJob(t, ctx, database, analyticsMaintenanceJob,
		analyticsLockExport, analyticsMaintenanceAttempt, "unknown", unknownAt)
	insertAnalyticsMaintenanceJob(t, ctx, database, analyticsMaintenanceDeadJob,
		analyticsMaintenanceDeadExport, analyticsMaintenanceAttemptID("dead"), "dead", unknownAt)
	insertAnalyticsMaintenanceJob(t, ctx, database, analyticsMaintenanceSuccessJob,
		analyticsMaintenanceSuccessExport, analyticsMaintenanceAttemptID("success"),
		"succeeded", unknownAt)
	insertAnalyticsMaintenanceJob(t, ctx, database, analyticsMaintenanceRunningJob,
		analyticsMaintenanceRunningExport, analyticsMaintenanceAttemptID("running"),
		"running", unknownAt)
	prepareAnalyticsMaintenanceTenantExport(t, ctx, database, exportUpdated, unknownAt)
	applyAnalyticsDocuments(t, ctx, database, "",
		analyticsMaintenancePlatformDocuments(t, "20260802020025", "20260802020049"))
	candidate := platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate{
		Kind:     platformasync.AnalyticsDashboardExportDeleteJobKind,
		TenantID: uuid.MustParse(analyticsTenantA), TenantSchema: "t_analytics_a",
		ExportID: uuid.MustParse(analyticsLockExport), AuthorityGroupID: "analytics",
		AuthorityGroupGeneration: 1, CutoverGeneration: 1,
		JobID:     uuid.MustParse(analyticsMaintenanceJob),
		AttemptID: uuid.MustParse(analyticsMaintenanceAttempt), AttemptNumber: 1,
		LeaseGeneration: 1, OriginalLeaseOwnerID: uuid.MustParse(analyticsMaintenanceOwner),
		EffectID:            analyticsMaintenanceEffectID(analyticsMaintenanceJob),
		EffectRequestSHA256: analyticsMaintenanceRequestSHA(), UnknownAt: unknownAt,
	}
	return analyticsMaintenanceProtocolFixture{database: database, candidate: candidate,
		observedAt: time.Now().UTC().Truncate(time.Microsecond)}
}

func analyticsMaintenancePlatformDocuments(t *testing.T,
	first, last string,
) []platformmigration.Document {
	t.Helper()
	documents, err := platformasync.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	already := map[string]bool{"20260802020001": true, "20260802020007": true,
		"20260802020014": true, "20260802020021": true}
	selected := make([]platformmigration.Document, 0, 24)
	for _, document := range documents {
		name := path.Base(document.Path())
		stamp := strings.SplitN(name, "_", 2)[0]
		if strings.Contains(document.Path(), "/fresh/") && stamp >= first && stamp <= last &&
			!already[stamp] {
			selected = append(selected, document)
		}
	}
	sort.Slice(selected, func(left, right int) bool {
		return selected[left].Path() < selected[right].Path()
	})
	return selected
}

func analyticsMaintenanceEffectID(seed string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("analytics-maintenance:"+seed))
}

func analyticsMaintenanceRequestSHA() (digest [32]byte) {
	for index := range digest {
		digest[index] = 0xa1
	}
	return digest
}

func analyticsMaintenanceAttemptID(seed string) string {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("attempt:"+seed)).String()
}

func insertAnalyticsMaintenanceJob(t *testing.T, ctx context.Context, database *sql.DB,
	jobID, exportID, attemptID, mode string, completedAt time.Time,
) {
	t.Helper()
	startedAt := completedAt.Add(-2 * time.Second)
	effectAt := completedAt.Add(-time.Second)
	owner := analyticsMaintenanceOwner
	effectID := analyticsMaintenanceEffectID(jobID)
	_, err := database.ExecContext(ctx, `WITH inserted_job AS (
INSERT INTO platform.jobs(
id,tenant_id,scope_kind,scope_id,authority_group_id,authority_group_generation,
authority_domain,cutover_generation,writer_origin,kind,payload_revision,payload_ref,
payload_sha256,dedupe_key,effect_class,effect_idempotency_key,priority,run_at,state,
attempts,max_attempts,lease_owner_id,lease_generation,lease_heartbeat_at,lease_until,
effect_started_at,last_error_class,last_error_code,last_error_sha256,retain_until,
created_at,updated_at,completed_at) VALUES($1,$2,'tenant',$2,'analytics',1,
'analytics',1,'v2','analytics.dashboard.export.delete',1,$3,
encode(sha256(convert_to($3,'UTF8')),'hex'),'analytics.dashboard.export.delete:'||$3,
'external','analytics.dashboard.export.delete:'||$3,100,$4::timestamptz,
CASE WHEN $5='running' THEN 'running' WHEN $5='succeeded' THEN 'succeeded' ELSE 'dead' END,
1,3,CASE WHEN $5='running' THEN $6::uuid END,1,
CASE WHEN $5='running' THEN $7::timestamptz END,
CASE WHEN $5='running' THEN $7::timestamptz+interval '5 minutes' END,
$7::timestamptz,CASE WHEN $5='dead' THEN 'permanent' END,
CASE WHEN $5='unknown' THEN 'JOB_OUTCOME_UNKNOWN' WHEN $5='dead' THEN 'DELETE_FAILED' END,
CASE WHEN $5='dead' THEN repeat('ab',32) END,
CASE WHEN $5='running' THEN NULL WHEN $5='succeeded' THEN $8::timestamptz+interval '90 days'
 ELSE $8::timestamptz+interval '365 days' END,$4::timestamptz,
CASE WHEN $5='running' THEN $7::timestamptz ELSE $8::timestamptz END,
CASE WHEN $5='running' THEN NULL ELSE $8::timestamptz END)
RETURNING id)
INSERT INTO platform.job_attempts(attempt_id,job_id,scope_kind,scope_id,attempt,
lease_generation,lease_owner_id,started_at,finished_at,outcome,error_class,error_code,
redacted_error_sha256,effect_started_at,effect_id,effect_request_sha256,effect_receipt_id)
SELECT $9,$1,'tenant',$2,1,1,$6,$4::timestamptz,
CASE WHEN $5='running' THEN NULL ELSE $8::timestamptz END,
CASE WHEN $5='running' THEN NULL WHEN $5='succeeded' THEN 'succeeded' ELSE 'dead' END,
CASE WHEN $5='dead' THEN 'permanent' END,
CASE WHEN $5='unknown' THEN 'JOB_OUTCOME_UNKNOWN' WHEN $5='dead' THEN 'DELETE_FAILED' END,
CASE WHEN $5='dead' THEN repeat('ab',32) END,$7::timestamptz,$10,repeat('a1',32),
NULL::uuid FROM inserted_job`, jobID, analyticsTenantA, exportID,
		startedAt, mode, owner, effectAt, completedAt, attemptID, effectID)
	if err != nil {
		t.Fatalf("insert Analytics maintenance %s job: %v", mode, err)
	}
}

func prepareAnalyticsMaintenanceTenantExport(t *testing.T, ctx context.Context,
	database *sql.DB, updatedAt, unknownAt time.Time,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback() }()
	_, err = transaction.ExecContext(ctx, `UPDATE t_analytics_a.analytics_dashboard_exports
SET state='cancelled',quota_state='released',delete_operation_id=$2,delete_job_id=$3,
delete_actor_id=actor_id,delete_idempotency_key='analytics-maintenance-delete',
delete_request_sha256=decode(repeat('d1',32),'hex'),row_revision=row_revision+1,
updated_at=$4 WHERE export_id=$1`, analyticsLockExport,
		analyticsLockOperation, analyticsMaintenanceJob, updatedAt.Add(time.Microsecond),
	)
	if err == nil {
		_, err = transaction.ExecContext(ctx, `UPDATE t_analytics_a.analytics_dashboard_exports
SET state='deleted',deleted_at=$2,row_revision=row_revision+1,updated_at=$2
WHERE export_id=$1`, analyticsLockExport, updatedAt.Add(2*time.Microsecond))
	}
	if err == nil {
		err = transaction.Commit()
	}
	if err != nil || !unknownAt.After(updatedAt.Add(2*time.Microsecond)) {
		t.Fatalf("prepare Analytics maintenance tenant export: %v", err)
	}
}
