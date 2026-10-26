package subscriptionapp

import (
	"context"
	"errors"
	"testing"
)

func TestResumeCancellationUsesCASAndReplacesCancellationJob(t *testing.T) {
	snapshot := fixtureSnapshot(t, "active")
	cancelAt := snapshot.Period.End()
	snapshot.CancelAtPeriodEnd, snapshot.CancelAt = true, &cancelAt
	job := fixtureJob(t, snapshot, BoundaryCancellation, cancelAt)
	store := &fakeStore{snapshot: snapshot, boundaryJobs: []JobIntent{job}}
	commands := fixtureCommands(t, store, fixtureNow)
	request := ResumeCancellationRequest{
		ActorID: snapshot.TenantID, SubscriptionID: snapshot.ID,
		SourceOperationID: fixtureV7(t), ExpectedRevision: snapshot.Revision,
	}
	result, err := commands.ResumeCancellation(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Subscription.CancelAtPeriodEnd || result.Subscription.CancelAt != nil ||
		result.Subscription.Revision != snapshot.Revision+1 {
		t.Fatalf("unexpected resumed snapshot: %+v", result.Subscription)
	}
	if len(store.boundaryJobs) != 1 || store.boundaryJobs[0].Boundary != BoundaryPeriodEnd ||
		store.boundaryJobs[0].ID == job.ID || !store.boundaryJobs[0].RunAt.Equal(cancelAt) {
		t.Fatalf("cancellation job was not atomically replaced: %+v", store.boundaryJobs)
	}
	if replay, err := commands.ResumeCancellation(context.Background(), request); err != nil || !replay.Replay || len(store.plans) != 1 {
		t.Fatalf("resume replay failed: result=%+v error=%v", replay, err)
	}
}

func TestResumeCancellationRejectsBoundaryAndStaleRevision(t *testing.T) {
	snapshot := fixtureSnapshot(t, "active")
	cancelAt := snapshot.Period.End()
	snapshot.CancelAtPeriodEnd, snapshot.CancelAt = true, &cancelAt
	job := fixtureJob(t, snapshot, BoundaryCancellation, cancelAt)
	request := ResumeCancellationRequest{
		ActorID: snapshot.TenantID, SubscriptionID: snapshot.ID,
		SourceOperationID: fixtureV7(t), ExpectedRevision: snapshot.Revision,
	}
	boundaryStore := &fakeStore{snapshot: snapshot, boundaryJobs: []JobIntent{job}}
	if _, err := fixtureCommands(t, boundaryStore, cancelAt).ResumeCancellation(
		context.Background(), request,
	); !errors.Is(err, ErrStateConflict) {
		t.Fatalf("boundary resume error = %v", err)
	}
	staleStore := &fakeStore{snapshot: snapshot, boundaryJobs: []JobIntent{job}}
	request.ExpectedRevision--
	if _, err := fixtureCommands(t, staleStore, fixtureNow).ResumeCancellation(
		context.Background(), request,
	); !errors.Is(err, ErrStaleRevision) {
		t.Fatalf("stale resume error = %v", err)
	}
}
