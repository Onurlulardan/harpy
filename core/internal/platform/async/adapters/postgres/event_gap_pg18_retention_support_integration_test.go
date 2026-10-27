//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/google/uuid"
)

func seedExpiredNormalOutbox(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) uuid.UUID {
	t.Helper()
	normal := *fixture
	normal.aggregate = uuid.New()
	normal.envelopes = make(map[int64]eventsv1.Envelope)
	envelope := normal.seedDeliveredAt(
		t, ctx, 1, time.Now().UTC().Add(-32*24*time.Hour),
	)
	return uuid.MustParse(envelope.EventID)
}

func insertExpiredNormalDeadLetter(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	liveRootID uuid.UUID,
	normalEventID uuid.UUID,
) uuid.UUID {
	t.Helper()
	normalRootID := uuid.New()
	const statement = `
INSERT INTO platform.dead_letters (
 id,root_id,entry_sequence,record_kind,tenant_id,scope_kind,scope_id,
 authority_group_id,authority_group_generation,authority_domain,
 cutover_generation,writer_origin,source_kind,source_id,consumer_name,
 consumer_version,source_contract,source_version,source_sha256,
 reason_class,reason_code,redacted_error_sha256,payload_canonical,
 payload_sha256,redacted_envelope,full_payload_expires_at,
 metadata_expires_at,recorded_at
)
SELECT $2,$2,0,'root',root.tenant_id,root.scope_kind,root.scope_id,
 root.authority_group_id,root.authority_group_generation,root.authority_domain,
 root.cutover_generation,root.writer_origin,root.source_kind,$3,
 root.consumer_name,root.consumer_version,root.source_contract,
 root.source_version,root.source_sha256,'permanent','EVENT_PERMANENT_FAILURE',
 root.redacted_error_sha256,root.payload_canonical,root.payload_sha256,
 root.redacted_envelope,pg_catalog.clock_timestamp()-interval '31 days',
 pg_catalog.clock_timestamp()+interval '365 days',
 pg_catalog.clock_timestamp()-interval '32 days'
FROM platform.dead_letters AS root WHERE root.id=$1`
	if _, err := fixture.runtime.Exec(
		ctx, statement, liveRootID, normalRootID, normalEventID,
	); err != nil {
		t.Fatalf("insert expired normal dead-letter: %v", err)
	}
	return normalRootID
}

func ageGapRetentionRows(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	rootID uuid.UUID,
	liveEventID uuid.UUID,
	normalEventID uuid.UUID,
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
		t.Fatal(err)
	}
	ids := []uuid.UUID{liveEventID}
	if normalEventID != uuid.Nil {
		ids = append(ids, normalEventID)
	}
	for _, eventID := range ids {
		if _, err := transaction.ExecContext(ctx, `
UPDATE platform.outbox SET
 confirmed_at=occurred_at+interval '1 hour',
 delivered_at=occurred_at+interval '1 hour'
WHERE event_id=$1`, eventID); err != nil {
			t.Fatalf("age outbox %s: %v", eventID, err)
		}
	}
	if _, err := transaction.ExecContext(
		ctx, "ALTER TABLE platform.outbox ENABLE ALWAYS TRIGGER outbox_transition_guard",
	); err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.ExecContext(
		ctx, "ALTER TABLE platform.dead_letters DISABLE TRIGGER dead_letters_redaction_guard",
	); err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.ExecContext(ctx, `
UPDATE platform.dead_letters
SET full_payload_expires_at=recorded_at+interval '1 microsecond'
WHERE id=$1`, rootID); err != nil {
		t.Fatalf("expire live gap root: %v", err)
	}
	if _, err := transaction.ExecContext(
		ctx, "ALTER TABLE platform.dead_letters ENABLE ALWAYS TRIGGER dead_letters_redaction_guard",
	); err != nil {
		t.Fatal(err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func runGapSafeRedaction(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) (int, int) {
	t.Helper()
	assertGapRetentionRuntimeIdentity(t, ctx, fixture)
	counts, err := runRetentionCycleCounts(ctx, fixture.runtime)
	if err != nil {
		t.Fatalf("run gap-safe retention cycle: %v", err)
	}
	return int(counts[1]), int(counts[2])
}

func assertGapRetentionRuntimeIdentity(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) {
	t.Helper()
	var current, session string
	if err := fixture.runtime.QueryRow(
		ctx, "SELECT current_user,session_user",
	).Scan(&current, &session); err != nil ||
		current != "harpy_runtime" || session != "harpy_v2_runtime" {
		t.Fatalf("gap retention identity = %s/%s: %v", current, session, err)
	}
}
