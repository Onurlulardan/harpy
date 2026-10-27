package subscriptionapp

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestExecuteDueBoundaryAppliesCancellationAndCompletesJobAtomically(t *testing.T) {
	snapshot := fixtureSnapshot(t, "active")
	dueAt := snapshot.Period.End()
	snapshot.CancelAtPeriodEnd, snapshot.CancelAt = true, &dueAt
	job := fixtureJob(t, snapshot, BoundaryCancellation, dueAt)
	boundary := BoundarySnapshot{Kind: BoundaryCancellation, DueAt: dueAt}
	store := &fakeStore{
		snapshot: snapshot, boundaryJobs: []JobIntent{job}, boundary: &boundary,
	}
	commands := fixtureCommands(t, store, dueAt)
	request := ExecuteDueBoundaryRequest{
		SubscriptionID: snapshot.ID, JobID: job.ID, EffectID: job.EffectID,
		ExpectedRevision: snapshot.Revision, LeaseGeneration: 3,
		Boundary: BoundaryCancellation, DueAt: dueAt,
	}
	result, err := commands.ExecuteDueBoundary(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Subscription.State != "cancelled" || result.Subscription.CancelAtPeriodEnd ||
		result.Subscription.CancelAt != nil || len(store.boundaryJobs) != 0 ||
		store.plans[0].CompleteJob == nil || store.plans[0].CompleteJob.ID != job.ID {
		t.Fatalf("cancellation boundary result=%+v jobs=%+v", result, store.boundaryJobs)
	}
	if replay, err := commands.ExecuteDueBoundary(context.Background(), request); err != nil || !replay.Replay || len(store.plans) != 1 {
		t.Fatalf("boundary replay failed: result=%+v error=%v", replay, err)
	}
}

func TestExecuteDueBoundaryRejectsEarlyAndStaleJob(t *testing.T) {
	snapshot := fixtureSnapshot(t, "active")
	dueAt := snapshot.Period.End()
	snapshot.CancelAtPeriodEnd, snapshot.CancelAt = true, &dueAt
	job := fixtureJob(t, snapshot, BoundaryCancellation, dueAt)
	boundary := BoundarySnapshot{Kind: BoundaryCancellation, DueAt: dueAt}
	request := ExecuteDueBoundaryRequest{
		SubscriptionID: snapshot.ID, JobID: job.ID, EffectID: job.EffectID,
		ExpectedRevision: snapshot.Revision, LeaseGeneration: 2,
		Boundary: BoundaryCancellation, DueAt: dueAt,
	}
	earlyStore := &fakeStore{
		snapshot: snapshot, boundaryJobs: []JobIntent{job}, boundary: &boundary,
	}
	if _, err := fixtureCommands(t, earlyStore, dueAt.Add(-time.Microsecond)).ExecuteDueBoundary(
		context.Background(), request,
	); !errors.Is(err, ErrBoundaryNotDue) {
		t.Fatalf("early boundary error = %v", err)
	}
	staleStore := &fakeStore{
		snapshot: snapshot, boundaryJobs: []JobIntent{job}, boundary: &boundary,
	}
	request.ExpectedRevision--
	if _, err := fixtureCommands(t, staleStore, dueAt).ExecuteDueBoundary(
		context.Background(), request,
	); !errors.Is(err, ErrStaleRevision) {
		t.Fatalf("stale boundary error = %v", err)
	}
}
