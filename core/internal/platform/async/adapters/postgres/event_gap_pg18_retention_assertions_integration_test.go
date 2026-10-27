//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func assertRuntimeCannotRemoveGapEvidence(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	rootID uuid.UUID,
) {
	t.Helper()
	if _, err := fixture.runtime.Exec(
		ctx, "DELETE FROM platform.dead_letters WHERE id=$1", rootID,
	); err == nil {
		t.Fatal("runtime deleted live event gap evidence")
	}
	if _, err := fixture.runtime.Exec(ctx, `
UPDATE platform.dead_letters
SET payload_canonical=NULL,payload_redacted_at=pg_catalog.clock_timestamp()
WHERE id=$1`, rootID); err == nil {
		t.Fatal("runtime redacted live event gap evidence")
	}
}

func assertRetentionPayloads(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	liveRootID uuid.UUID,
	normalRootID uuid.UUID,
	liveEventID uuid.UUID,
	normalEventID uuid.UUID,
	liveProtected bool,
) {
	t.Helper()
	var liveRoot, normalRoot, liveEvent, normalEvent bool
	err := fixture.owner.QueryRowContext(ctx, `
SELECT
 (SELECT payload_canonical IS NOT NULL FROM platform.dead_letters WHERE id=$1),
 (SELECT payload_canonical IS NULL AND payload_redacted_at IS NOT NULL
  FROM platform.dead_letters WHERE id=$2),
 (SELECT payload_canonical IS NOT NULL FROM platform.outbox WHERE event_id=$3),
 (SELECT payload_canonical IS NULL AND payload_redacted_at IS NOT NULL
  FROM platform.outbox WHERE event_id=$4)`,
		liveRootID, normalRootID, liveEventID, normalEventID,
	).Scan(&liveRoot, &normalRoot, &liveEvent, &normalEvent)
	if err != nil || normalRoot != true || normalEvent != true ||
		liveRoot != liveProtected || liveEvent != liveProtected {
		t.Fatalf("retention payloads live root/event=%t/%t normal=%t/%t: %v",
			liveRoot, liveEvent, normalRoot, normalEvent, err)
	}
}
