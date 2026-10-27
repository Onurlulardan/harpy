//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func installGapReconcileCommitCrash(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	jobID uuid.UUID,
) {
	t.Helper()
	const statement = `
CREATE TABLE platform.gap_reconcile_crash_target (job_id uuid PRIMARY KEY);
CREATE FUNCTION platform.gap_reconcile_commit_crash_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
BEGIN
 IF NEW.state='succeeded' AND EXISTS (
  SELECT 1 FROM platform.gap_reconcile_crash_target AS target
  WHERE target.job_id=NEW.id
 ) THEN
  RAISE EXCEPTION USING ERRCODE='P5698',
   MESSAGE='injected event gap reconcile commit crash';
 END IF;
 RETURN NEW;
END;
$function$;
CREATE CONSTRAINT TRIGGER gap_reconcile_commit_crash
AFTER UPDATE ON platform.jobs DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION platform.gap_reconcile_commit_crash_v1();`
	if _, err := fixture.owner.ExecContext(ctx, statement); err != nil {
		t.Fatalf("install gap reconcile commit crash: %v", err)
	}
	if _, err := fixture.owner.ExecContext(ctx,
		"INSERT INTO platform.gap_reconcile_crash_target VALUES ($1)", jobID,
	); err != nil {
		t.Fatalf("arm gap reconcile commit crash: %v", err)
	}
}

func removeGapReconcileCommitCrash(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) {
	t.Helper()
	const statement = `
DROP TRIGGER gap_reconcile_commit_crash ON platform.jobs;
DROP FUNCTION platform.gap_reconcile_commit_crash_v1();
DROP TABLE platform.gap_reconcile_crash_target;`
	if _, err := fixture.owner.ExecContext(ctx, statement); err != nil {
		t.Fatalf("remove gap reconcile commit crash: %v", err)
	}
}

func expireGapJobLease(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	jobID uuid.UUID,
) {
	t.Helper()
	transaction, err := fixture.owner.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err := transaction.ExecContext(
		ctx, "ALTER TABLE platform.jobs DISABLE TRIGGER jobs_transition_guard",
	); err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.ExecContext(ctx, `
UPDATE platform.jobs SET
 lease_heartbeat_at=pg_catalog.clock_timestamp()-interval '61 seconds',
 lease_until=pg_catalog.clock_timestamp()-interval '1 second',
 updated_at=pg_catalog.clock_timestamp()
WHERE id=$1`, jobID); err != nil {
		t.Fatalf("expire gap job lease: %v", err)
	}
	if _, err := transaction.ExecContext(
		ctx, "ALTER TABLE platform.jobs ENABLE ALWAYS TRIGGER jobs_transition_guard",
	); err != nil {
		t.Fatal(err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func eventGapAuthorityForStore(
	t *testing.T,
	ctx context.Context,
	store *workerRepository,
	tenantID uuid.UUID,
) JobAuthority {
	t.Helper()
	token, err := store.BeginJobAuthorityPage(256)
	if err != nil {
		t.Fatal(err)
	}
	page, err := store.ListJobAuthorities(ctx, token)
	if err != nil {
		t.Fatal(err)
	}
	for _, authority := range page.Authorities() {
		if authority.scopeID == tenantID && authority.groupID == "workflow" &&
			authority.domain == "workflow" {
			return authority
		}
	}
	t.Fatal("restarted event gap authority was not listed")
	return JobAuthority{}
}

func assertGapReconcileRolledBack(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	jobID uuid.UUID,
) {
	t.Helper()
	var state, delivery string
	var actions int
	err := fixture.owner.QueryRowContext(ctx, `
SELECT job.state,event.delivery_state,
 (SELECT count(*) FROM platform.dead_letters WHERE id=job.id)
FROM platform.jobs AS job
JOIN platform.dead_letters AS root
 ON job.payload_ref='deadletter:event-gap:'||root.id::text
JOIN platform.outbox AS event ON event.event_id=root.source_id
WHERE job.id=$1`, jobID).Scan(&state, &delivery, &actions)
	if err != nil || state != "running" || delivery != "delivered" || actions != 0 {
		t.Fatalf("reconcile rollback state/delivery/actions = %s/%s/%d: %v",
			state, delivery, actions, err)
	}
}
