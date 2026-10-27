//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"database/sql"
	"path"
	"sort"
	"strings"
	"testing"
	"time"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformmigration "github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

const (
	analyticsLockExport    = "019b2000-0000-7000-8000-000000000061"
	analyticsLockJob       = "019b2000-0000-7000-8000-000000000062"
	analyticsLockExpiryJob = "019b2000-0000-7000-8000-000000000063"
	analyticsLockLease     = "019b2000-0000-7000-8000-000000000064"
	analyticsLockOperation = "019b2000-0000-7000-8000-000000000065"
	analyticsLockDeleteJob = "019b2000-0000-7000-8000-000000000066"
)

func prepareAnalyticsPlatformProtocolPG18(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()
	database := startAnalyticsPG18(t, ctx, "platform_lock_order")
	prepareAnalyticsPrerequisites(t, ctx, database, "fresh")
	applyAnalyticsDocuments(t, ctx, database, "", analyticsDocuments(t, "fresh", "platform"))
	reconciled, remaining := reconcileAnalyticsAuthorities(t, ctx, database, true)
	if reconciled != 2 || remaining != 0 {
		t.Fatalf("Analytics authority reconcile = %d/%d, want 2/0", reconciled, remaining)
	}
	for _, schema := range []string{"t_analytics_a", "t_analytics_b"} {
		applyAnalyticsDocuments(t, ctx, database, schema,
			analyticsDocuments(t, "fresh", "tenant"))
	}
	applyAnalyticsDocuments(t, ctx, database, "",
		analyticsPlatformProtocolDocuments(t))
	assertAnalyticsTenantIsolation(t, ctx, database)
	assertAnalyticsQueryPublishing(t, ctx, database)
	assertAnalyticsDashboardPublishing(t, ctx, database)
	createdAt := time.Now().UTC().Add(250 * time.Millisecond)
	submitAnalyticsPlatformLockExport(t, ctx, database, createdAt)
	insertAnalyticsPlatformLockJob(t, ctx, database, createdAt)
	return database
}

func analyticsPlatformProtocolDocuments(t *testing.T) []platformmigration.Document {
	t.Helper()
	wanted := map[string]bool{
		"20260715020000_create_async_reject_mutation_routine.fresh.sql":                     true,
		"20260715020100_create_jobs.fresh.sql":                                              true,
		"20260715020200_create_job_attempts.fresh.sql":                                      true,
		"20260715020300_create_outbox.fresh.sql":                                            true,
		"20260715020400_create_inbox.fresh.sql":                                             true,
		"20260715020500_create_idempotency_receipts.fresh.sql":                              true,
		"20260715020600_create_dead_letters.fresh.sql":                                      true,
		"20260715020700_create_async_authority_assertion_v1.fresh.sql":                      true,
		"20260715020800_create_async_append_only_guards_v1.fresh.sql":                       true,
		"20260802020001_create_analytics_dashboard_export_claim_v1.fresh.sql":               true,
		"20260802020007_create_analytics_dashboard_export_unknown_resolutions_v1.fresh.sql": true,
		"20260802020014_create_analytics_dashboard_export_cleanup_owner_v1.fresh.sql":       true,
		"20260802020021_create_analytics_dashboard_export_control_v1.fresh.sql":             true,
	}
	documents, err := platformasync.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	selected := make([]platformmigration.Document, 0, len(wanted))
	for _, document := range documents {
		name := path.Base(document.Path())
		if wanted[name] && strings.Contains(document.Path(), "/fresh/") {
			selected = append(selected, document)
			delete(wanted, name)
		}
	}
	if len(wanted) != 0 {
		t.Fatalf("Platform Async protocol fixture migrations missing: %v", wanted)
	}
	sort.Slice(selected, func(left, right int) bool {
		return selected[left].Path() < selected[right].Path()
	})
	return selected
}

func submitAnalyticsPlatformLockExport(t *testing.T, ctx context.Context,
	database *sql.DB, occurredAt time.Time,
) {
	t.Helper()
	connection := analyticsRuntimeConnection(t, ctx, database)
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var exportID, state string
	err := transaction.QueryRowContext(ctx, `SELECT export_id,effective_state
FROM analytics_dashboard_export_submit_v2('t_analytics_a',$1,$2,$3,$4,$5,$6,$7,$8,
'csv','{}'::bytea,1,1,1,100000,67108864,$9,$10)`,
		"019b2000-0000-7000-8000-000000000060", analyticsDashboardActor,
		"analytics-lock-order-export", bytes.Repeat([]byte{0xc1}, 32),
		analyticsLockExport, analyticsLockJob, analyticsLockExpiryJob,
		analyticsDashboard, occurredAt, occurredAt.Add(24*time.Hour)).Scan(&exportID, &state)
	if err != nil || exportID != analyticsLockExport || state != "queued" {
		t.Fatalf("submit lock-order export = %s/%s: %v", exportID, state, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func insertAnalyticsPlatformLockJob(t *testing.T, ctx context.Context,
	database *sql.DB, occurredAt time.Time,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `INSERT INTO platform.jobs(
id,tenant_id,scope_kind,scope_id,authority_group_id,authority_group_generation,
authority_domain,cutover_generation,writer_origin,kind,payload_revision,payload_ref,
payload_sha256,dedupe_key,effect_class,effect_idempotency_key,priority,run_at,state,
attempts,max_attempts,lease_generation,created_at,updated_at)
VALUES($1,$2,'tenant',$2,'analytics',1,'analytics',1,'v2',
'analytics.dashboard.export.render',1,$3,encode(sha256(convert_to($3,'UTF8')),'hex'),
'analytics.dashboard.export.render:'||$3,'external',
'analytics.dashboard.export.render:'||$3,100,$4,'queued',0,3,0,$4,$4)`,
		analyticsLockJob, analyticsTenantA, analyticsLockExport,
		occurredAt.Add(-time.Second))
	if err != nil {
		t.Fatalf("insert lock-order Platform job: %v", err)
	}
}
