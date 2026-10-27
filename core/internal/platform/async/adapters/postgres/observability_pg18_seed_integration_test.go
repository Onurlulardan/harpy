//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func seedObservabilityJob(
	t *testing.T, ctx context.Context, fixture publisherPGFixture,
	tenantID uuid.UUID, runAt time.Time,
) {
	t.Helper()
	jobID := uuid.New()
	arguments := []any{
		jobID, tenantID, "tenant", tenantID,
		"operations", int64(1), "workflow", int64(1), "v2",
		"observability.job", 1, "payload/jobs/" + jobID.String(),
		asyncSHA256([]byte("observability job payload")), "observe-" + jobID.String(),
		"database_transactional", nil, 0, runAt.UTC().Truncate(time.Microsecond), 3, nil,
	}
	result, err := fixture.runtime.Exec(ctx, enqueueJobSQL, arguments...)
	if err != nil {
		t.Fatalf("seed observability job: %v", err)
	}
	if result.RowsAffected() != 1 {
		t.Fatalf("seed observability job affected %d rows", result.RowsAffected())
	}
}

func ageOutboxForObservation(
	t *testing.T, ctx context.Context, fixture publisherPGFixture,
	eventID uuid.UUID, ageSeconds int,
) {
	t.Helper()
	transaction, err := fixture.owner.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err := transaction.ExecContext(
		ctx, "ALTER TABLE platform.outbox DISABLE TRIGGER outbox_transition_guard",
	); err != nil {
		t.Fatalf("disable outbox guard for clock fixture: %v", err)
	}
	result, err := transaction.ExecContext(ctx, `
UPDATE platform.outbox
SET created_at=pg_catalog.clock_timestamp()-($2::integer*interval '1 second'),
    occurred_at=pg_catalog.clock_timestamp()-(($2::integer+1)*interval '1 second')
WHERE event_id=$1`, eventID, ageSeconds)
	if err != nil {
		t.Fatalf("age outbox clock fixture: %v", err)
	}
	if affected, err := result.RowsAffected(); err != nil || affected != 1 {
		t.Fatalf("age outbox affected %d rows: %v", affected, err)
	}
	if _, err := transaction.ExecContext(
		ctx, "ALTER TABLE platform.outbox ENABLE ALWAYS TRIGGER outbox_transition_guard",
	); err != nil {
		t.Fatalf("enable outbox guard after clock fixture: %v", err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func assertObservabilityRuntimeRole(
	t *testing.T, ctx context.Context, fixture publisherPGFixture,
) {
	t.Helper()
	var current, session string
	if err := fixture.runtime.QueryRow(ctx, "SELECT current_user,session_user").Scan(
		&current, &session,
	); err != nil {
		t.Fatal(err)
	}
	if current != "harpy_runtime" || session != "harpy_v2_runtime" {
		t.Fatalf("async observer identities = %s/%s", current, session)
	}
}
