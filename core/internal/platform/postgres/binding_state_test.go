package postgres

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBindingSealRejectsQueuedAndFutureOperations(t *testing.T) {
	state, transaction := bindingFixture()
	active, err := state.acquire(context.Background(), transaction)
	if err != nil {
		t.Fatal(err)
	}
	active.installAbort(active.finish)
	queuedResult := make(chan error, 1)
	go func() {
		queued, acquireErr := state.acquire(context.Background(), transaction)
		if queued != nil {
			queued.finish()
		}
		queuedResult <- acquireErr
	}()
	waitBindingOperations(t, state, 2)
	sealResult := make(chan error, 1)
	go func() { sealResult <- state.seal(time.Second) }()
	waitBindingClosed(t, state)
	if operation, err := state.acquire(context.Background(), transaction); operation != nil ||
		!errors.Is(err, ErrTransactionClosed) {
		t.Fatalf("post-seal acquisition = %v, %v", operation, err)
	}
	if err := <-queuedResult; !errors.Is(err, ErrTransactionClosed) {
		t.Fatalf("queued acquisition error = %v", err)
	}
	if err := <-sealResult; !errors.Is(err, ErrBindingAbandoned) {
		t.Fatalf("seal result = %v", err)
	}
}

func TestBindingSealUsesExactBoundedCleanupWait(t *testing.T) {
	state, transaction := bindingFixture()
	operation, err := state.acquire(context.Background(), transaction)
	if err != nil {
		t.Fatal(err)
	}
	const cleanup = 30 * time.Millisecond
	started := time.Now()
	err = state.seal(cleanup)
	elapsed := time.Since(started)
	operation.finish()
	if !errors.Is(err, ErrBindingCleanupTimeout) {
		t.Fatalf("cleanup timeout error = %v", err)
	}
	if elapsed < cleanup || elapsed > 250*time.Millisecond {
		t.Fatalf("cleanup wait = %v, want %v bounded", elapsed, cleanup)
	}
}

func TestBindingContextCancellationReleasesActiveResource(t *testing.T) {
	state, transaction := bindingFixture()
	ctx, cancel := context.WithCancel(context.Background())
	operation, err := state.acquire(ctx, transaction)
	if err != nil {
		t.Fatal(err)
	}
	operation.installAbort(operation.finish)
	cancel()
	waitBindingOperations(t, state, 0)
	next, err := state.acquire(context.Background(), transaction)
	if err != nil {
		t.Fatalf("operation gate remained locked after cancellation: %v", err)
	}
	next.finish()
}

func bindingFixture() (*bindingState, *RuntimeTransaction) {
	state := newBindingState()
	transaction := &RuntimeTransaction{bindings: state, cleanup: time.Second}
	transaction.open.Store(true)
	return state, transaction
}

func waitBindingOperations(t *testing.T, state *bindingState, count int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		state.mutex.Lock()
		current := len(state.operations)
		state.mutex.Unlock()
		if current == count {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("binding operation count did not become %d", count)
}

func waitBindingClosed(t *testing.T, state *bindingState) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if !state.accepting.Load() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("binding state did not stop accepting")
}
