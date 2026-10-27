//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func assertTerminalGapRetryEvidenceLive(
	t *testing.T,
	ctx context.Context,
	value terminalGapRetryCase,
	retryJobID uuid.UUID,
) {
	t.Helper()
	var state string
	var retryOf uuid.UUID
	var rootLive, eventLive bool
	var actions int
	err := value.fixture.runtime.QueryRow(ctx, `
SELECT retry.state,retry.retry_of_job_id,
 root.payload_canonical IS NOT NULL AND root.payload_redacted_at IS NULL,
 event.payload_canonical IS NOT NULL AND event.payload_redacted_at IS NULL,
 (SELECT count(*) FROM platform.dead_letters AS action
  WHERE action.root_id=$4 AND action.record_kind='retry_action')
FROM platform.jobs AS retry
JOIN platform.dead_letters AS root ON root.id=$2
JOIN platform.outbox AS event ON event.event_id=$3
WHERE retry.id=$1`, retryJobID, value.rootID, value.sourceEventID,
		value.terminalRoot,
	).Scan(&state, &retryOf, &rootLive, &eventLive, &actions)
	if err != nil || state != "queued" || retryOf != value.jobID ||
		!rootLive || !eventLive || actions != 1 {
		t.Fatalf(
			"committed gap retry = %s/%s evidence %t/%t actions %d: %v",
			state, retryOf, rootLive, eventLive, actions, err,
		)
	}
}

func assertTerminalGapRootRedacted(
	t *testing.T,
	ctx context.Context,
	value terminalGapRetryCase,
) {
	t.Helper()
	var redacted bool
	err := value.fixture.runtime.QueryRow(ctx, `
SELECT payload_canonical IS NULL AND payload_redacted_at IS NOT NULL
FROM platform.dead_letters WHERE id=$1`, value.rootID).Scan(&redacted)
	if err != nil || !redacted {
		t.Fatalf("redaction-winning gap root state = %t: %v", redacted, err)
	}
}
