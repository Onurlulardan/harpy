package providerapp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type finalizerStoreStub struct {
	called bool
	result func(FinalizeCommandMutation) Finalization
}

func (store *finalizerStoreStub) FinalizeCommand(
	_ context.Context,
	mutation FinalizeCommandMutation,
) (Finalization, error) {
	store.called = true
	if store.result != nil {
		return store.result(mutation), nil
	}
	return Finalization{}, nil
}

func TestFinalizeCommandAcceptsOnlyDecisionStatePairs(t *testing.T) {
	commandID := uuid.MustParse("018f1000-0000-7000-8000-000000000071")
	request := FinalizeCommandRequest{
		TenantID: uuid.New(), JobID: uuid.MustParse(
			"018f1000-0000-7000-8000-000000000072",
		), AttemptID: uuid.New(), CommandID: commandID,
		LeaseGeneration: 1, PayloadRevision: 1,
		PayloadReference: "billing-provider-finalize/" + commandID.String(),
		PayloadSHA256:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	pairs := []struct {
		decision FinalizeDecision
		state    CommandState
	}{
		{FinalizePending, CommandQueued},
		{FinalizeSucceeded, CommandSucceeded},
		{FinalizeFailed, CommandFailed},
		{FinalizeReconcile, CommandReconcileRequired},
		{FinalizeReconcileUnknown, CommandReconcileRequired},
		{FinalizeReplay, CommandSucceeded},
	}
	for _, pair := range pairs {
		t.Run(string(pair.decision), func(t *testing.T) {
			store := &finalizerStoreStub{result: func(
				mutation FinalizeCommandMutation,
			) Finalization {
				activeID := uuid.MustParse(
					"018f1000-0000-7000-8000-000000000073",
				)
				if pair.decision == FinalizeReconcile {
					activeID = mutation.ReconciliationJobID
				}
				finalizerID := request.JobID
				if pair.decision == FinalizePending || pair.decision == FinalizeReconcile {
					finalizerID = mutation.NextFinalizerJobID
				}
				return Finalization{
					Decision: pair.decision, CommandID: commandID,
					ActorID: uuid.New(), ActiveJobID: activeID,
					FinalizerJobID: finalizerID,
					Provider:       "manual", RequestSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					State: pair.state, Revision: 2, OccurredAt: now,
				}
			}}
			finalizer, err := NewCommandFinalizer(
				store, uuid.NewV7, func() time.Time { return now },
			)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := finalizer.FinalizeCommand(
				context.Background(), request,
			); err != nil {
				t.Fatalf("FinalizeCommand() error = %v", err)
			}
		})
	}
}

func TestFinalizeCommandRejectsMismatchedDecisionState(t *testing.T) {
	commandID := uuid.MustParse("018f1000-0000-7000-8000-000000000081")
	now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	store := &finalizerStoreStub{result: func(
		FinalizeCommandMutation,
	) Finalization {
		return Finalization{
			Decision: FinalizeSucceeded, CommandID: commandID,
			ActorID: uuid.New(), ActiveJobID: uuid.MustParse(
				"018f1000-0000-7000-8000-000000000082",
			), FinalizerJobID: uuid.MustParse(
				"018f1000-0000-7000-8000-000000000083",
			), Provider: "manual",
			RequestSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			State:         CommandFailed, Revision: 2, OccurredAt: now,
		}
	}}
	finalizer, err := NewCommandFinalizer(
		store, uuid.NewV7, func() time.Time { return now },
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = finalizer.FinalizeCommand(
		context.Background(), FinalizeCommandRequest{
			TenantID: uuid.New(), JobID: uuid.MustParse(
				"018f1000-0000-7000-8000-000000000083",
			), AttemptID: uuid.New(), CommandID: commandID,
			LeaseGeneration: 1, PayloadRevision: 1,
			PayloadReference: "billing-provider-finalize/" + commandID.String(),
			PayloadSHA256:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
	)
	if !errors.Is(err, ErrCommandState) {
		t.Fatalf("FinalizeCommand() error = %v", err)
	}
}

func TestFinalizeCommandRejectsInvalidClaimBeforeStorage(t *testing.T) {
	store := &finalizerStoreStub{}
	finalizer, err := NewCommandFinalizer(store, uuid.NewV7, func() time.Time {
		return time.Now().UTC().Truncate(time.Microsecond)
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = finalizer.FinalizeCommand(
		context.Background(), FinalizeCommandRequest{},
	)
	if !errors.Is(err, ErrCommandRequest) || store.called {
		t.Fatalf("FinalizeCommand() error = %v, store called = %v", err, store.called)
	}
}
