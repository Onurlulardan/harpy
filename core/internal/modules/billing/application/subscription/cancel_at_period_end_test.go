package subscriptionapp

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestCancelAtPeriodEndCommitsScheduleReceiptOutboxAndJobAtomically(t *testing.T) {
	snapshot := fixtureSnapshot(t, "active")
	store := &fakeStore{snapshot: snapshot}
	commands := fixtureCommands(t, store, fixtureNow)
	request := CancelAtPeriodEndRequest{
		ActorID: snapshot.TenantID, SubscriptionID: snapshot.ID,
		SourceOperationID: fixtureV7(t), ExpectedRevision: snapshot.Revision,
		Reason: "tenant requested cancellation",
	}
	result, err := commands.CancelAtPeriodEnd(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Replay || !result.Subscription.CancelAtPeriodEnd ||
		result.Subscription.CancelAt == nil || result.Subscription.Revision != snapshot.Revision+1 {
		t.Fatalf("unexpected cancellation result: %+v", result)
	}
	if len(store.receipts) != 1 || len(store.outbox) != 1 || len(store.audits) != 1 ||
		len(store.boundaryJobs) != 1 || store.boundaryJobs[0].Boundary != BoundaryCancellation {
		t.Fatalf("atomic effects missing: receipts=%d outbox=%d audit=%d jobs=%d",
			len(store.receipts), len(store.outbox), len(store.audits), len(store.boundaryJobs))
	}

	replay, err := commands.CancelAtPeriodEnd(context.Background(), request)
	if err != nil || !replay.Replay || replay.Subscription.Revision != result.Subscription.Revision ||
		len(store.plans) != 1 {
		t.Fatalf("idempotent replay failed: result=%+v error=%v plans=%d", replay, err, len(store.plans))
	}

	reused := request
	reused.Reason = "different request"
	if _, err := commands.CancelAtPeriodEnd(context.Background(), reused); !errors.Is(err, ErrIdempotencyReused) {
		t.Fatalf("reused operation error = %v", err)
	}
}

func TestCancelAtPeriodEndPreservesEmptyAndTrimmedReasons(t *testing.T) {
	for _, test := range []struct {
		name, input, want string
	}{
		{"empty", "", ""},
		{"trimmed", "  requested by tenant  ", "requested by tenant"},
	} {
		t.Run(test.name, func(t *testing.T) {
			snapshot := fixtureSnapshot(t, "active")
			store := &fakeStore{snapshot: snapshot}
			commands := fixtureCommands(t, store, fixtureNow)
			_, err := commands.CancelAtPeriodEnd(context.Background(), CancelAtPeriodEndRequest{
				ActorID: snapshot.TenantID, SubscriptionID: snapshot.ID,
				SourceOperationID: fixtureV7(t), ExpectedRevision: snapshot.Revision,
				Reason: test.input,
			})
			if err != nil || len(store.audits) != 1 || store.audits[0].Reason != test.want {
				t.Fatalf("reason=%q audits=%+v error=%v", test.want, store.audits, err)
			}
		})
	}
}

func TestCancelAtPeriodEndSerializesConcurrentReplayAndStaleRevision(t *testing.T) {
	snapshot := fixtureSnapshot(t, "active")
	store := &fakeStore{snapshot: snapshot}
	commands := fixtureCommands(t, store, fixtureNow)
	request := CancelAtPeriodEndRequest{
		ActorID: snapshot.TenantID, SubscriptionID: snapshot.ID,
		SourceOperationID: fixtureV7(t), ExpectedRevision: snapshot.Revision,
		Reason: "cancel",
	}
	results := make(chan Result, 2)
	errorsFound := make(chan error, 2)
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, err := commands.CancelAtPeriodEnd(context.Background(), request)
			results <- result
			errorsFound <- err
		}()
	}
	wait.Wait()
	close(results)
	close(errorsFound)
	replays := 0
	for err := range errorsFound {
		if err != nil {
			t.Fatal(err)
		}
	}
	for result := range results {
		if result.Replay {
			replays++
		}
	}
	if replays != 1 || len(store.plans) != 1 || store.snapshot.Revision != snapshot.Revision+1 {
		t.Fatalf("replays=%d plans=%d revision=%d", replays, len(store.plans), store.snapshot.Revision)
	}

	secondStore := &fakeStore{snapshot: snapshot}
	secondCommands := fixtureCommands(t, secondStore, fixtureNow)
	requests := []CancelAtPeriodEndRequest{request, request}
	requests[1].SourceOperationID = fixtureV7(t)
	errorsFound = make(chan error, 2)
	for _, candidate := range requests {
		wait.Add(1)
		go func(value CancelAtPeriodEndRequest) {
			defer wait.Done()
			_, err := secondCommands.CancelAtPeriodEnd(context.Background(), value)
			errorsFound <- err
		}(candidate)
	}
	wait.Wait()
	close(errorsFound)
	stale, success := 0, 0
	for err := range errorsFound {
		if err == nil {
			success++
		} else if errors.Is(err, ErrStaleRevision) {
			stale++
		}
	}
	if success != 1 || stale != 1 || len(secondStore.plans) != 1 {
		t.Fatalf("success=%d stale=%d plans=%d", success, stale, len(secondStore.plans))
	}
}

func TestCancelAtPeriodEndRollsBackEveryEffectOnCommitFailure(t *testing.T) {
	snapshot := fixtureSnapshot(t, "active")
	store := &fakeStore{snapshot: snapshot, failCommit: true}
	commands := fixtureCommands(t, store, fixtureNow)
	_, err := commands.CancelAtPeriodEnd(context.Background(), CancelAtPeriodEndRequest{
		ActorID: snapshot.TenantID, SubscriptionID: snapshot.ID,
		SourceOperationID: fixtureV7(t), ExpectedRevision: snapshot.Revision, Reason: "cancel",
	})
	if !errors.Is(err, errFakeCommit) || store.snapshot.Revision != snapshot.Revision ||
		len(store.receipts) != 0 || len(store.outbox) != 0 || len(store.audits) != 0 ||
		len(store.boundaryJobs) != 0 {
		t.Fatalf("rollback failed: error=%v store=%+v", err, store)
	}
}
