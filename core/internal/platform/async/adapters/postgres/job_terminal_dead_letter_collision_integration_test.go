//go:build integration

package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func assertTerminalJobRootCollisionFailsClosed(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
) {
	t.Helper()
	jobID, claim := claimLifecycleDatabaseJob(t, ctx, fixture, 8)
	const code = "WORKFLOW_COLLISION"
	rootID := terminalJobRootID(jobID, claim.Attempt(), code)
	insertEventDeadLetterRoot(
		t, ctx, fixture, rootID, uuid.New(), "workflow.collision",
	)
	failure := terminalJobFailure(t, jobsv1.ErrorPermanent, code, "9")
	err := commitBoundFailure(ctx, fixture, claim, failure)
	var databaseError *pgconn.PgError
	if !errors.As(err, &databaseError) || databaseError.Code != "P5683" {
		t.Fatalf("unrelated root collision error=%v", err)
	}
	state, open := lifecycleState(t, ctx, fixture, jobID)
	if state != string(jobsv1.StateRunning) || open != 1 ||
		terminalJobRootCount(t, ctx, fixture, jobID) != 0 {
		t.Fatalf("collision rollback state=%s open=%d", state, open)
	}
	var collisionRootCount int
	if err := fixture.owner.QueryRowContext(
		ctx, "SELECT count(*) FROM platform.dead_letters WHERE id=$1", rootID,
	).Scan(&collisionRootCount); err != nil || collisionRootCount != 1 {
		t.Fatalf("collision evidence count=%d error=%v", collisionRootCount, err)
	}
}

func assertEventRootsRemainPerConsumer(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
) {
	t.Helper()
	eventID := uuid.New()
	insertEventDeadLetterRoot(
		t, ctx, fixture, uuid.New(), eventID, "workflow.projector",
	)
	insertEventDeadLetterRoot(
		t, ctx, fixture, uuid.New(), eventID, "analytics.projector",
	)
	var roots int
	if err := fixture.owner.QueryRowContext(ctx, `
SELECT count(*) FROM platform.dead_letters
WHERE source_kind='event' AND record_kind='root' AND source_id=$1`, eventID).Scan(&roots); err != nil || roots != 2 {
		t.Fatalf("per-consumer event roots=%d error=%v", roots, err)
	}
}

func insertEventDeadLetterRoot(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	rootID uuid.UUID,
	eventID uuid.UUID,
	consumer string,
) {
	t.Helper()
	payload := []byte(fmt.Sprintf(`{"event_id":%q}`, eventID.String()))
	_, err := fixture.runtime.Exec(ctx, `
WITH evidence_clock AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS now
)
INSERT INTO platform.dead_letters (
 id,root_id,entry_sequence,record_kind,tenant_id,scope_kind,scope_id,
 authority_group_id,authority_group_generation,authority_domain,
 cutover_generation,writer_origin,source_kind,source_id,consumer_name,
 consumer_version,source_contract,source_version,source_sha256,
 reason_class,reason_code,redacted_error_sha256,payload_canonical,
 payload_sha256,redacted_envelope,full_payload_expires_at,
 metadata_expires_at,recorded_at
)
SELECT $1,$1,0,'root',$2,'tenant',$2,'operations',1,'workflow',1,'v2',
 'event',$3::uuid,$4::text,1,'forms.entry.created',1,repeat('b',64),
 'permanent','EVENT_REJECTED',repeat('c',64),$5,$6,
 pg_catalog.jsonb_build_object('event_id',$3::uuid,'consumer',$4::text),
 now+interval '30 days',now+interval '365 days',now
FROM evidence_clock`,
		rootID, asyncTenantA, eventID, consumer,
		payload, deadLetterDigest(payload),
	)
	if err != nil {
		t.Fatalf("insert event dead-letter root: %v", err)
	}
}
