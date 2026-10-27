//go:build integration

package postgres

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
	"github.com/google/uuid"
)

func TestEventGapLifecycleOnPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
			defer cancel()
			fixture := newEventGapPGFixture(t, ctx, origin)
			assertEventGapReadinessChain(t, ctx, &fixture)
		})
	}
}

func assertEventGapReadinessChain(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) {
	t.Helper()
	for sequence := int64(1); sequence <= 4; sequence++ {
		fixture.seedDelivered(t, ctx, sequence)
	}
	if fixture.deliver(t, ctx, 3) != brokerrabbit.HandleCommitted ||
		fixture.deliver(t, ctx, 4) != brokerrabbit.HandleCommitted {
		t.Fatal("durably parked gaps were not reported as committed transactions")
	}
	root3, job3 := fixture.gapIDs(t, ctx, 3)
	_, job4 := fixture.gapIDs(t, ctx, 4)
	if fixture.deliver(t, ctx, 4) != brokerrabbit.HandleCommitted {
		t.Fatal("duplicate gap did not commit deterministic evidence")
	}
	assertSingleGapEvidence(t, ctx, fixture, root3, job3)
	assertNormalJobReadiness(t, ctx, fixture)
	if fixture.gapReady(t, ctx, job3) || fixture.gapReady(t, ctx, job4) {
		t.Fatal("gap became ready without a predecessor watermark")
	}
	assertOtherTenantCannotAdvanceGap(t, ctx, fixture, job4)
	fixture.deliver(t, ctx, 1)
	if fixture.gapReady(t, ctx, job3) {
		t.Fatal("E3 became ready at watermark 1")
	}
	fixture.deliver(t, ctx, 2)
	if !fixture.gapReady(t, ctx, job3) || fixture.gapReady(t, ctx, job4) {
		t.Fatal("watermark 2 did not make only E3 ready")
	}
	claim3 := fixture.claimGap(t, ctx, job3)
	assertConcurrentGapReconcile(t, ctx, fixture, claim3)
	assertGapRequeueEvidence(t, ctx, fixture, job3, fixture.envelopes[3].EventID)
	republishGap(t, ctx, fixture, 3)
	if !fixture.gapReady(t, ctx, job4) {
		t.Fatal("watermark 3 did not make E4 ready")
	}
	claim4 := fixture.claimGap(t, ctx, job4)
	result, err := fixture.store.ReconcileEventGap(ctx, claim4)
	if err != nil || !result.Requeued() || result.ActionID() != job4 {
		t.Fatalf("reconcile E4 = %#v, %v", result, err)
	}
	republishGap(t, ctx, fixture, 4)
	assertGapEffects(t, ctx, fixture, 4)
}

func assertConcurrentGapReconcile(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	claim JobClaim,
) {
	t.Helper()
	errorsSeen := make(chan error, 2)
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, err := fixture.store.ReconcileEventGap(ctx, claim)
			errorsSeen <- err
		}()
	}
	wait.Wait()
	close(errorsSeen)
	succeeded, fenced := 0, 0
	for err := range errorsSeen {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, ErrJobLeaseLost):
			fenced++
		default:
			t.Fatalf("concurrent gap reconcile error = %v", err)
		}
	}
	if succeeded != 1 || fenced != 1 {
		t.Fatalf("concurrent reconciliation success/fenced = %d/%d", succeeded, fenced)
	}
}

func republishGap(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	sequence int64,
) {
	t.Helper()
	envelope := fixture.envelopes[sequence]
	authority := publisherAuthorityFor(t, ctx, fixture.publisherPGFixture, fixture.tenantID)
	claim := claimPipelineEvent(
		t, ctx, fixture.publisherPGFixture, authority,
		uuid.MustParse(envelope.EventID), 2,
	)
	outcome, err := fixture.handler.Handle(
		ctx, claim.Envelope(), brokerrabbit.DeliveryMetadata{Redelivered: true},
	)
	if err != nil || outcome != brokerrabbit.HandleCommitted {
		t.Fatalf("replayed event %d outcome = %s, %v", sequence, outcome, err)
	}
	if _, err := fixture.store.ConfirmPublish(
		ctx, claim, eventGapDigest("replay:"+envelope.EventID),
	); err != nil {
		t.Fatalf("confirm replayed event %d: %v", sequence, err)
	}
}

func assertNormalJobReadiness(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) {
	t.Helper()
	jobID := uuid.New()
	_, err := fixture.runtime.Exec(ctx, enqueueJobSQL,
		jobID, fixture.tenantID, "tenant", fixture.tenantID, "workflow", int64(1),
		"workflow", int64(1), "v2", "workflow.normal", 1,
		"payload/"+jobID.String(), eventGapDigest("payload"), "normal:"+jobID.String(),
		string(jobsv1.EffectDatabaseTransactional), nil, 1, time.Now().UTC(), 1, nil,
	)
	if err != nil || !fixture.gapReady(t, ctx, jobID) {
		t.Fatalf("normal job fast-path ready = %v", err)
	}
}
