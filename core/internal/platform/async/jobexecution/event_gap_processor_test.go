package jobexecution

import (
	"context"
	"errors"
	"testing"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

type eventGapProcessorStore struct {
	calls int
	err   error
}

func (store *eventGapProcessorStore) ReconcileEventGap(
	context.Context,
	asyncpostgres.JobClaim,
) (asyncpostgres.EventGapReconciliation, error) {
	store.calls++
	return asyncpostgres.EventGapReconciliation{}, store.err
}

func TestEventGapProcessorRequiresCanonicalPolicyAndStore(t *testing.T) {
	t.Parallel()
	policy, err := platformasync.EventGapJobPolicy("workflow")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewEventGapProcessor(policy, nil); !errors.Is(
		err, ErrEventGapProcessorInvalid,
	) {
		t.Fatalf("nil store error = %v", err)
	}
	drifted := policy
	drifted.MaximumAttempts--
	if _, err := NewEventGapProcessor(drifted, &eventGapProcessorStore{}); !errors.Is(
		err, ErrEventGapProcessorInvalid,
	) {
		t.Fatalf("drifted policy error = %v", err)
	}
}

func TestEventGapProcessorDelegatesOnlyToDatabaseStore(t *testing.T) {
	t.Parallel()
	policy, _ := platformasync.EventGapJobPolicy("workflow")
	want := errors.New("database unavailable")
	store := &eventGapProcessorStore{err: want}
	processor, err := NewEventGapProcessor(policy, store)
	if err != nil || processor.Policy() != policy {
		t.Fatalf("new processor = %#v, %v", processor, err)
	}
	if err := processor.Process(context.Background(), asyncpostgres.JobClaim{}); !errors.Is(
		err, want,
	) || store.calls != 1 {
		t.Fatalf("process error = %v, calls = %d", err, store.calls)
	}
	if err := processor.Process(nil, asyncpostgres.JobClaim{}); !errors.Is(
		err, ErrEventGapProcessorInvalid,
	) {
		t.Fatalf("nil context error = %v", err)
	}
}
