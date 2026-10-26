//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertRetentionCycleWork(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	ids retentionCycleWorkIDs,
) {
	t.Helper()
	var checks [10]bool
	err := database.QueryRowContext(ctx, `
SELECT
 (SELECT payload_ref IS NULL AND payload_redacted_at IS NOT NULL
    FROM platform.jobs WHERE id=$1),
 NOT EXISTS (SELECT 1 FROM platform.jobs WHERE id=$2),
 NOT EXISTS (SELECT 1 FROM platform.inbox
             WHERE record_kind='event_receipt' AND event_id=$3),
 NOT EXISTS (SELECT 1 FROM platform.idempotency_receipts WHERE receipt_id=$4),
 (SELECT payload_canonical IS NULL AND payload_redacted_at IS NOT NULL
    FROM platform.outbox WHERE event_id=$5),
 (SELECT payload_canonical IS NULL AND payload_redacted_at IS NOT NULL
    FROM platform.dead_letters WHERE id=$6),
 NOT EXISTS (SELECT 1 FROM platform.dead_letters WHERE id=$7),
 (SELECT payload_canonical IS NOT NULL AND payload_redacted_at IS NULL
    FROM platform.outbox WHERE event_id=$8),
 (SELECT payload_canonical IS NOT NULL AND payload_redacted_at IS NULL
    FROM platform.dead_letters WHERE id=$9),
 (SELECT payload_canonical IS NOT NULL AND payload_redacted_at IS NULL
    FROM platform.dead_letters WHERE id=$10)`,
		ids.redactJob, ids.purgeJob, ids.purgeInbox, ids.purgeReceipt,
		ids.normalEvent, ids.normalRoot, ids.purgeRoot, ids.liveEvent,
		ids.liveRoot, ids.staleRoot,
	).Scan(
		&checks[0], &checks[1], &checks[2], &checks[3], &checks[4],
		&checks[5], &checks[6], &checks[7], &checks[8], &checks[9],
	)
	if err != nil {
		t.Fatalf("read retention cycle effects: %v", err)
	}
	labels := [...]string{
		"job redaction", "job purge", "inbox purge", "receipt purge",
		"outbox redaction", "dead-letter redaction", "dead-letter purge",
		"live-gap outbox protection", "live-gap root protection",
		"stale-job root protection",
	}
	for index, passed := range checks {
		if !passed {
			t.Errorf("retention cycle check failed: %s", labels[index])
		}
	}
}
