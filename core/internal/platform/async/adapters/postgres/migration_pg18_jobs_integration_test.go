//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertAsyncJobMatrices(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	assertAsyncParkGenerationMatrix(t, ctx, database)
	assertAsyncRetentionMatrix(t, ctx, database)
}

func assertAsyncParkGenerationMatrix(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	cases := []struct {
		name             string
		groupGeneration  int
		domainGeneration int
		accepted         bool
	}{
		{name: "group advanced", groupGeneration: 2, domainGeneration: 1, accepted: true},
		{name: "domain advanced", groupGeneration: 1, domainGeneration: 2, accepted: true},
		{name: "both current", groupGeneration: 1, domainGeneration: 1},
		{name: "older proof", groupGeneration: 0, domainGeneration: 2},
	}
	for _, test := range cases {
		t.Run("park_"+test.name, func(t *testing.T) {
			withAsyncRolledBackTransaction(t, ctx, database, func(tx *sql.Tx) {
				id := "81000000-0000-4000-8000-000000000001"
				if err := insertAsyncQueuedJob(
					ctx, tx, id, asyncTenantA, "tenant", asyncTenantA,
					"operations", "workflow", "park-job",
				); err != nil {
					t.Fatalf("insert queued job: %v", err)
				}
				_, err := tx.ExecContext(ctx, `
UPDATE platform.jobs
SET parked_at = TIMESTAMPTZ '2026-07-15 09:30:00+00',
    parked_code = 'JOB_AUTHORITY_STALE',
    parked_current_group_generation = $2,
    parked_current_cutover_generation = $3
WHERE id = $1`, id, test.groupGeneration, test.domainGeneration)
				if test.accepted && err != nil {
					t.Fatalf("valid park proof rejected: %v", err)
				}
				if !test.accepted {
					requireAsyncConstraintRejection(t, err)
				}
			})
		})
	}
}

func assertAsyncRetentionMatrix(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	cases := []struct {
		name     string
		update   string
		accepted bool
	}{
		{name: "queued has no retention", update: `retain_until = TIMESTAMPTZ '2026-10-15 10:00:00+00'`},
		{name: "succeeded exact floor", accepted: true, update: asyncSucceededRetention("90 days")},
		{name: "succeeded below floor", update: asyncSucceededRetention("89 days")},
		{name: "dead exact floor", accepted: true, update: asyncDeadRetention("365 days")},
		{name: "dead below floor", update: asyncDeadRetention("364 days")},
		{name: "release proof complete", accepted: true, update: asyncReleaseProof(true)},
		{name: "release proof partial", update: asyncReleaseProof(false)},
	}
	for _, test := range cases {
		t.Run("retention_"+test.name, func(t *testing.T) {
			withAsyncRolledBackTransaction(t, ctx, database, func(tx *sql.Tx) {
				id := "81000000-0000-4000-8000-000000000002"
				if err := insertAsyncQueuedJob(
					ctx, tx, id, asyncTenantA, "tenant", asyncTenantA,
					"operations", "workflow", "retention-job",
				); err != nil {
					t.Fatalf("insert queued job: %v", err)
				}
				_, err := tx.ExecContext(ctx,
					"UPDATE platform.jobs SET "+test.update+" WHERE id = $1", id)
				if test.accepted && err != nil {
					t.Fatalf("valid retention proof rejected: %v", err)
				}
				if !test.accepted {
					requireAsyncConstraintRejection(t, err)
				}
			})
		})
	}
}

func asyncSucceededRetention(interval string) string {
	return `state = 'succeeded', updated_at = TIMESTAMPTZ '2026-07-15 10:00:00+00',
completed_at = TIMESTAMPTZ '2026-07-15 10:00:00+00',
retain_until = TIMESTAMPTZ '2026-07-15 10:00:00+00' + INTERVAL '` + interval + `'`
}

func asyncDeadRetention(interval string) string {
	return `state = 'dead', updated_at = TIMESTAMPTZ '2026-07-15 10:00:00+00',
completed_at = TIMESTAMPTZ '2026-07-15 10:00:00+00',
retain_until = TIMESTAMPTZ '2026-07-15 10:00:00+00' + INTERVAL '` + interval + `',
last_error_class = 'permanent', last_error_code = 'JOB_FAILED',
last_error_sha256 = repeat('7', 64)`
}

func asyncReleaseProof(complete bool) string {
	proof := `retention_release_revision = 1`
	if complete {
		proof += `, retention_release_sha256 = repeat('8', 64),
retention_released_at = TIMESTAMPTZ '2026-07-15 10:00:00+00'`
	}
	return asyncSucceededRetention("90 days") + ", " + proof
}
