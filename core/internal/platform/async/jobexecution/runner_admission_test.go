package jobexecution

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
)

type admissionStore struct {
	*runnerTestStore
	reconciled chan struct{}
	once       sync.Once
}

func (store *admissionStore) ReconcileStaleJobs(context.Context, int) (int, error) {
	store.once.Do(func() { close(store.reconciled) })
	return 0, nil
}

func TestRunnerDoesNotTouchStoreUntilAdmissionOpens(t *testing.T) {
	store := &admissionStore{
		runnerTestStore: &runnerTestStore{listed: make(chan struct{})},
		reconciled:      make(chan struct{}),
	}
	registry, err := NewRegistry(nil)
	if err != nil {
		t.Fatal(err)
	}
	gate := admission.NewGate(nil)
	runner, err := NewRunner(store, registry, gate)
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-store.reconciled:
		t.Fatal("closed admission allowed job store access")
	case <-time.After(2 * idlePollInterval):
	}
	if !gate.Open() {
		t.Fatal("job admission did not open")
	}
	select {
	case <-store.listed:
	case <-time.After(time.Second):
		t.Fatal("opened admission did not resume job scanning")
	}
	runner.BeginDrain()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runner.Wait(ctx); err != nil {
		t.Fatal(err)
	}
}
