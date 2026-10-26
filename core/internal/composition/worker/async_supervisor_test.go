package worker

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type supervisorWorkload struct {
	startErr error
	failures chan error
	done     chan struct{}
	stop     sync.Once
	started  bool
	drained  bool
	forced   bool
}

func newSupervisorWorkload() *supervisorWorkload {
	return &supervisorWorkload{failures: make(chan error, 1), done: make(chan struct{})}
}

func (workload *supervisorWorkload) Start(context.Context) error {
	workload.started = true
	return workload.startErr
}

func (workload *supervisorWorkload) BeginDrain() {
	workload.drained = true
	workload.finish()
}

func (workload *supervisorWorkload) Wait(ctx context.Context) error {
	select {
	case <-workload.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (workload *supervisorWorkload) Force() error {
	workload.forced = true
	workload.finish()
	return nil
}

func (workload *supervisorWorkload) Failures() <-chan error { return workload.failures }

func (workload *supervisorWorkload) finish() {
	workload.stop.Do(func() {
		close(workload.done)
		close(workload.failures)
	})
}

func TestAsyncSupervisorStartsDrainsAndWaitsForEveryWorkload(t *testing.T) {
	first, second := newSupervisorWorkload(), newSupervisorWorkload()
	supervisor, err := newAsyncSupervisor(first, second)
	if err != nil {
		t.Fatal(err)
	}
	if err := supervisor.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	supervisor.BeginDrain()
	waitContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := supervisor.Wait(waitContext); err != nil {
		t.Fatal(err)
	}
	if !first.started || !second.started || !first.drained || !second.drained {
		t.Fatalf("workloads = %#v, %#v", first, second)
	}
	if failure, open := <-supervisor.Failures(); open || failure != nil {
		t.Fatalf("supervisor failure = %v, open = %v", failure, open)
	}
}

func TestAsyncSupervisorForwardsFailureAndRollsBackPartialStart(t *testing.T) {
	want := errors.New("workload failed")
	first, second := newSupervisorWorkload(), newSupervisorWorkload()
	supervisor, _ := newAsyncSupervisor(first, second)
	if err := supervisor.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	first.failures <- want
	first.finish()
	second.finish()
	if failure := <-supervisor.Failures(); !errors.Is(failure, want) {
		t.Fatalf("forwarded failure = %v", failure)
	}

	started := newSupervisorWorkload()
	rejected := newSupervisorWorkload()
	rejected.startErr = want
	rollback, _ := newAsyncSupervisor(started, rejected)
	if err := rollback.Start(context.Background()); !errors.Is(err, want) {
		t.Fatalf("partial start error = %v", err)
	}
	if !started.forced {
		t.Fatal("partial start did not force the active workload")
	}
}

func TestAsyncSupervisorRejectsIncompleteInputs(t *testing.T) {
	if _, err := newAsyncSupervisor(); err == nil {
		t.Fatal("empty supervisor was accepted")
	}
	if _, err := newAsyncSupervisor(nil); err == nil {
		t.Fatal("nil workload was accepted")
	}
	valid, _ := newAsyncSupervisor(newSupervisorWorkload())
	if err := valid.Start(nil); err == nil {
		t.Fatal("nil start context was accepted")
	}
	if err := valid.Wait(nil); err == nil {
		t.Fatal("nil wait context was accepted")
	}
}
