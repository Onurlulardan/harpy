//go:build integration

package postgres

import (
	"context"
	"testing"
)

func assertGapRaceHasNoPartialCommit(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	eventID string,
) {
	t.Helper()
	var roots int
	var payloadPresent bool
	err := fixture.owner.QueryRowContext(ctx, `
SELECT
 (SELECT count(*) FROM platform.dead_letters
  WHERE source_id=$1 AND reason_class='sequence_gap'),
 (SELECT payload_canonical IS NOT NULL AND payload_redacted_at IS NULL
  FROM platform.outbox WHERE event_id=$1)`, eventID,
	).Scan(&roots, &payloadPresent)
	if err != nil || roots != 0 || !payloadPresent {
		t.Fatalf("gap race exposed partial commit = roots %d payload %t: %v",
			roots, payloadPresent, err)
	}
}

func assertGapRaceEvidence(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	eventID string,
) {
	t.Helper()
	var roots, jobs int
	var payloadPresent bool
	err := fixture.owner.QueryRowContext(ctx, `
SELECT
 (SELECT count(*) FROM platform.dead_letters
  WHERE source_id=$1 AND reason_class='sequence_gap'),
 (SELECT count(*) FROM platform.jobs AS job JOIN platform.dead_letters AS root
  ON job.payload_ref='deadletter:event-gap:'||root.id::text
  WHERE root.source_id=$1),
 (SELECT payload_canonical IS NOT NULL AND payload_redacted_at IS NULL
  FROM platform.outbox WHERE event_id=$1)`, eventID,
	).Scan(&roots, &jobs, &payloadPresent)
	if err != nil || roots != 1 || jobs != 1 || !payloadPresent {
		t.Fatalf("gap race evidence = roots %d jobs %d payload %t: %v",
			roots, jobs, payloadPresent, err)
	}
}
