//go:build integration

package postgres

import (
	"context"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
	"github.com/google/uuid"
)

func assertSingleGapEvidence(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	rootID uuid.UUID,
	jobID uuid.UUID,
) {
	t.Helper()
	var roots, jobs int
	err := fixture.owner.QueryRowContext(ctx, `
SELECT
 (SELECT count(*) FROM platform.dead_letters
  WHERE id=$1 AND root_id=$1 AND record_kind='root'
    AND reason_class='sequence_gap' AND payload_canonical IS NOT NULL),
 (SELECT count(*) FROM platform.jobs
  WHERE id=$2 AND payload_ref='deadletter:event-gap:'||$1::text
    AND state='queued' AND attempts=0)`, rootID, jobID).Scan(&roots, &jobs)
	if err != nil || roots != 1 || jobs != 1 {
		t.Fatalf("deterministic gap root/job = %d/%d: %v", roots, jobs, err)
	}
}

func assertOtherTenantCannotAdvanceGap(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	jobID uuid.UUID,
) {
	t.Helper()
	if _, err := fixture.owner.ExecContext(ctx, `
CREATE TABLE pipeline_tenant_b.pipeline_effect_probe (event_id uuid PRIMARY KEY);
GRANT SELECT,INSERT ON pipeline_tenant_b.pipeline_effect_probe TO harpy_runtime;`); err != nil {
		t.Fatalf("prepare second tenant effect table: %v", err)
	}
	other := *fixture
	other.tenantID = uuid.MustParse(asyncTenantB)
	other.envelopes = make(map[int64]eventsv1.Envelope)
	other.handler = newPipelineConsumerHandlerForTenant(
		t, fixture.worker.coordinator, fixture.catalog, fixture.consumers, other.tenantID,
	)
	for sequence := int64(1); sequence <= 3; sequence++ {
		other.seedDelivered(t, ctx, sequence)
		if outcome := other.deliver(t, ctx, sequence); outcome != brokerrabbit.HandleCommitted {
			t.Fatalf("second tenant event %d outcome = %s", sequence, outcome)
		}
	}
	if fixture.gapReady(t, ctx, jobID) {
		t.Fatal("another tenant's watermark advanced the event gap")
	}
}

func assertGapRequeueEvidence(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	jobID uuid.UUID,
	eventID string,
) {
	t.Helper()
	var actions int
	var state, delivery string
	var exactClock bool
	err := fixture.owner.QueryRowContext(ctx, `
SELECT
 (SELECT count(*) FROM platform.dead_letters
  WHERE id=$1 AND record_kind='retry_action' AND source_id=$2),
 job.state,event.delivery_state,action.recorded_at=event.next_attempt_at
FROM platform.jobs AS job
JOIN platform.dead_letters AS action ON action.id=job.id
JOIN platform.outbox AS event ON event.event_id=action.source_id
WHERE job.id=$1`, jobID, eventID).Scan(&actions, &state, &delivery, &exactClock)
	if err != nil || actions != 1 || state != "succeeded" || delivery != "pending" || !exactClock {
		t.Fatalf("gap action/job/outbox = %d %s %s exact=%t: %v",
			actions, state, delivery, exactClock, err)
	}
}

func assertGapEffects(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	want int,
) {
	t.Helper()
	var effects, receipts, watermarks int
	var through int64
	err := fixture.owner.QueryRowContext(ctx, `
SELECT
 (SELECT count(*) FROM pipeline_tenant_a.pipeline_effect_probe),
 count(*) FILTER (WHERE record_kind='event_receipt'),
 count(*) FILTER (WHERE record_kind='watermark'),
 COALESCE(max(through_sequence),0)
FROM platform.inbox
WHERE consumer_name='projection.workflow' AND consumer_version=1
  AND scope_kind='tenant' AND scope_id=$1
  AND aggregate_type='workflow.instance' AND aggregate_id=$2`,
		fixture.tenantID, fixture.aggregate,
	).Scan(&effects, &receipts, &watermarks, &through)
	if err != nil || effects != want || receipts != want || watermarks != want || through != int64(want) {
		t.Fatalf("effect/receipt/watermark/through = %d/%d/%d/%d: %v",
			effects, receipts, watermarks, through, err)
	}
}
