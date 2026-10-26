package retention

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
)

type blockingStore struct{ entered chan struct{} }

func (store blockingStore) RunRetentionCycle(ctx context.Context) (CycleResult, error) {
	close(store.entered)
	<-ctx.Done()
	return CycleResult{}, ctx.Err()
}

func TestRunnerDrainCancelsActiveCycleAndClosesCleanly(t *testing.T) {
	runner, err := NewRunner(
		blockingStore{entered: make(chan struct{})}, &recordingObserver{},
		openRetentionAdmission(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	store := runner.store.(blockingStore)
	select {
	case <-store.entered:
	case <-time.After(time.Second):
		t.Fatal("retention cycle did not start")
	}
	runner.BeginDrain()
	waitContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runner.Wait(waitContext); err != nil {
		t.Fatal(err)
	}
	if failure, open := <-runner.Failures(); open || failure != nil {
		t.Fatalf("drained failure = %v, open=%t", failure, open)
	}
	if err := runner.Start(context.Background()); !errors.Is(err, ErrAlreadyStarted) {
		t.Fatalf("second start error = %v", err)
	}
}

func TestRunnerReportsStoreFailureAndValidatesLifecycle(t *testing.T) {
	want := errors.New("retention database failed")
	runner, _ := NewRunner(
		&sequenceStore{err: want}, &recordingObserver{}, openRetentionAdmission(t),
	)
	if err := runner.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	failure, open := <-runner.Failures()
	if !open || !errors.Is(failure, want) {
		t.Fatalf("reported failure = %v, open=%t", failure, open)
	}
	if _, open := <-runner.Failures(); open {
		t.Fatal("failure channel remained open")
	}
	if _, err := NewRunner(
		nil, &recordingObserver{}, openRetentionAdmission(t),
	); !errors.Is(err, ErrDependencies) {
		t.Fatalf("nil store error = %v", err)
	}
	if _, err := NewRunner(
		&sequenceStore{}, nil, openRetentionAdmission(t),
	); !errors.Is(err, ErrDependencies) {
		t.Fatalf("nil observer error = %v", err)
	}
	if _, err := NewRunner(&sequenceStore{}, &recordingObserver{}, nil); !errors.Is(err, ErrDependencies) {
		t.Fatalf("nil admission error = %v", err)
	}
	if err := runner.Wait(nil); !errors.Is(err, ErrDependencies) {
		t.Fatalf("nil wait error = %v", err)
	}
	if err := runner.Force(); err != nil {
		t.Fatal(err)
	}
}

func openRetentionAdmission(t *testing.T) *admission.Gate {
	t.Helper()
	gate := admission.NewGate(nil)
	if !gate.Open() {
		t.Fatal("retention admission did not open")
	}
	return gate
}
