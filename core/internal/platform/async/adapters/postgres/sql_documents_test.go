package postgres

import (
	"strings"
	"testing"
)

func TestWriterSQLDocumentsAreStaticAndConflictAware(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		source  string
		markers []string
	}{
		{"enqueue job", enqueueJobSQL, []string{
			"INSERT INTO platform.jobs", "ON CONFLICT DO NOTHING", "'queued', 0",
		}},
		{"find job", findJobEnqueueSQL, []string{
			"pg_catalog.count(*) = 1", "payload_redacted_at IS NOT NULL",
		}},
		{"enqueue event", enqueueEventSQL, []string{
			"INSERT INTO platform.outbox", "ON CONFLICT DO NOTHING", "'pending', 0",
		}},
		{"find event", findEventEnqueueSQL, []string{
			"pg_catalog.count(*) = 1", "payload_schema_sha256 = $20",
		}},
		{"park event gap", parkEventGapSQL, []string{
			"locked_source AS MATERIALIZED", "FOR UPDATE OF event",
			"event.payload_sha256 = gap.value->'envelope'->>'payload_sha256'",
			"pg_catalog.convert_from(event.payload_canonical, 'UTF8')::jsonb",
			"FROM gap_clock CROSS JOIN locked_source",
		}},
		{"list job authorities", listJobAuthoritiesSQL, []string{
			"JOIN platform.cutover_groups", "JOIN platform.domain_cutovers",
			"job.run_at<=job_clock.observed_at", "job.state='running'",
			"job.lease_until<=job_clock.observed_at", "LIMIT $6",
			"platform.cutover_job_bindings", "'cutover_shadow'::text",
		}},
		{"list stale jobs", listStaleJobsSQL, []string{
			"group_row.generation > job.authority_group_generation",
			"domain_row.generation > job.cutover_generation",
			"job.state IN ('queued', 'running')", "job.effect_started_at IS NULL",
			"ORDER BY job.created_at, job.id", "LIMIT $1",
		}},
		{"park stale job", parkStaleJobSQL, []string{
			"platform.platform_async_park_stale_work_v2", "$2::text", "$5::uuid",
		}},
		{"list stale external jobs", listStaleExternalJobsSQL, []string{
			"JOIN platform.job_attempts", "job.effect_class = 'external'",
			"job.effect_started_at IS NOT NULL", "attempt.finished_at IS NULL",
			"ORDER BY job.created_at, job.id", "LIMIT $1",
		}},
		{"finalize stale external job", finalizeStaleExternalJobSQL, []string{
			"platform.platform_async_finalize_stale_external_unknown_v1",
			"$3::bigint", "$5::timestamptz",
		}},
		{"assert job authority", assertJobAuthoritySQL, []string{
			"platform.platform_async_assert_worker_authority_v1", "$7::text",
		}},
		{"claim jobs", claimJobsSQL, []string{
			"FOR UPDATE SKIP LOCKED", "job.priority DESC, job.run_at, job.id",
			"interval '60 seconds'", "LIMIT $8", "pg_catalog.generate_subscripts(",
			"($13::integer[])[policy.position] = job.max_attempts",
			"$14::text = 'cutover_shadow'", "platform.cutover_job_bindings",
			"pg_catalog.left(",
			"platform.platform_async_event_gap_job_ready_v2(job.id)",
		}},
		{"reconcile event gap", reconcileEventGapSQL, []string{
			"platform.platform_async_reconcile_event_gap_v2($1, $2, $3, $4)",
			"completed_at, requeued, action_id, disposition",
		}},
		{"resolve stale disposition authority", resolveStaleJobAuthoritySQL, []string{
			"job.parked_code = 'JOB_AUTHORITY_STALE'", "job.state IN ('queued', 'cancelled')",
			"authority_group.generation >= job.parked_current_group_generation",
		}},
		{"inspect stale disposition", inspectStaleJobDispositionSQL, []string{
			"job.parked_at", "job.effect_started_at", "pg_catalog.transaction_timestamp()",
		}},
		{"dispose stale job", disposeStaleJobSQL, []string{
			"platform.platform_async_dispose_stale_job_v1", "$15::bytea", "$23::text",
		}},
		{"assert live job catalog", assertLiveJobCatalogSQL, []string{
			"job.state IN ('queued', 'running')", "pg_catalog.generate_subscripts(",
			"($5::integer[])[policy.position] = job.max_attempts",
		}},
		{"insert job attempts", insertJobAttemptsSQL, []string{
			"INSERT INTO platform.job_attempts", "pg_catalog.uuidv7()",
			"$7::timestamptz[]", "RETURNING attempt_id, job_id",
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, marker := range test.markers {
				if !strings.Contains(test.source, marker) {
					t.Fatalf("SQL lacks %q", marker)
				}
			}
			if strings.Contains(test.source, "%!") || strings.Contains(test.source, "{{") {
				t.Fatal("SQL contains a runtime formatting surface")
			}
		})
	}
	if strings.Contains(resolveStaleJobAuthoritySQL, "FOR SHARE") ||
		strings.Contains(resolveStaleJobAuthoritySQL, "FOR UPDATE") {
		t.Fatal("stale authority resolver must remain a SELECT-only snapshot")
	}
}
