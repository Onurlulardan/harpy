//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func advancePublisherAuthority(
	t *testing.T, ctx context.Context, fixture publisherPGFixture,
	tenantID uuid.UUID, generation int64,
) {
	t.Helper()
	transaction, err := fixture.owner.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	for name, statement := range map[string]string{
		"group": `UPDATE platform.cutover_groups SET generation=$1
            WHERE scope_kind='tenant' AND scope_id=$2 AND group_id='operations'`,
		"domain": `UPDATE platform.domain_cutovers SET generation=$1,
            changed_at=pg_catalog.clock_timestamp()
            WHERE scope_kind='tenant' AND scope_id=$2 AND domain='workflow'`,
	} {
		result, err := transaction.ExecContext(ctx, statement, generation, tenantID)
		if err != nil {
			t.Fatalf("advance publisher %s authority: %v", name, err)
		}
		if affected, err := result.RowsAffected(); err != nil || affected != 1 {
			t.Fatalf("advance publisher %s affected %d: %v", name, affected, err)
		}
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func assertStaleOutboxEvidence(
	t *testing.T, ctx context.Context, fixture publisherPGFixture,
	firstID, secondID uuid.UUID,
) {
	t.Helper()
	var parkedRows, totalRows int64
	if err := fixture.owner.QueryRowContext(ctx, `
SELECT count(*) FILTER (WHERE
        delivery_state='pending' AND parked_at IS NOT NULL AND
        parked_code='OUTBOX_AUTHORITY_STALE' AND
        parked_current_group_generation=2 AND
        parked_current_cutover_generation=2 AND
        lease_owner_id IS NULL AND lease_heartbeat_at IS NULL AND lease_until IS NULL),
       count(*)
FROM platform.outbox WHERE event_id IN ($1,$2)`, firstID, secondID).Scan(
		&parkedRows, &totalRows,
	); err != nil {
		t.Fatal(err)
	}
	if parkedRows != 2 || totalRows != 2 {
		t.Fatalf("parked outbox rows = %d/%d, want 2/2", parkedRows, totalRows)
	}
	var evidenceRows, validRows int64
	if err := fixture.owner.QueryRowContext(ctx, `
SELECT count(*), count(*) FILTER (WHERE
        dead.record_kind='root' AND dead.source_kind='event' AND
        dead.consumer_name='platform.outbox.publisher' AND
        dead.reason_class='stale_generation' AND
        dead.reason_code='OUTBOX_AUTHORITY_STALE' AND
        dead.authority_group_generation=2 AND dead.cutover_generation=2 AND
        dead.payload_canonical=event.payload_canonical AND
        dead.payload_sha256=event.payload_sha256 AND
        (dead.redacted_envelope->>'stored_group_generation')::bigint=1 AND
        (dead.redacted_envelope->>'stored_domain_generation')::bigint=1 AND
        (dead.redacted_envelope->>'current_group_generation')::bigint=2 AND
        (dead.redacted_envelope->>'current_domain_generation')::bigint=2)
FROM platform.dead_letters AS dead
JOIN platform.outbox AS event ON event.event_id=dead.source_id
WHERE dead.source_id IN ($1,$2)`, firstID, secondID).Scan(
		&evidenceRows, &validRows,
	); err != nil {
		t.Fatal(err)
	}
	if evidenceRows != 2 || validRows != 2 {
		t.Fatalf("stale DLQ evidence = %d/%d, want 2/2", validRows, evidenceRows)
	}
	_, err := fixture.runtime.Exec(ctx, `UPDATE platform.dead_letters
        SET reason_code=reason_code WHERE source_id=$1`, firstID)
	assertPermissionDenied(t, err)
}
