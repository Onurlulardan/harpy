package dashboardexport

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestWorkloadRenderStartFailureTerminatesRecoveryAndIsOneShot(t *testing.T) {
	t.Parallel()
	recovery := newAsyncWorkloadFixture()
	render := newAsyncWorkloadFixture()
	render.startErr = ErrUnavailable
	workload, err := NewWorkload(render, recovery)
	if err != nil {
		t.Fatal(err)
	}
	if err = workload.Start(context.Background()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Start() error = %v", err)
	}
	if !recovery.forced || !recovery.waited {
		t.Fatalf("partial start recovery forced=%v waited=%v",
			recovery.forced, recovery.waited)
	}
	if _, open := <-workload.Failures(); open {
		t.Fatal("failure channel remained open after partial start rollback")
	}
	if err = workload.Start(context.Background()); err == nil {
		t.Fatal("partially started workload was reusable")
	}
}

func TestWorkloadRenderStartFailureJoinsRecoveryShutdownErrors(t *testing.T) {
	t.Parallel()
	recovery := newAsyncWorkloadFixture()
	forceErr, waitErr := errors.New("force recovery"), errors.New("wait recovery")
	recovery.forceErr, recovery.waitErr = forceErr, waitErr
	render := newAsyncWorkloadFixture()
	render.startErr = ErrUnavailable
	workload, err := NewWorkload(render, recovery)
	if err != nil {
		t.Fatal(err)
	}
	err = workload.Start(context.Background())
	for _, want := range []error{ErrUnavailable, forceErr, waitErr} {
		if !errors.Is(err, want) {
			t.Fatalf("Start() error = %v; want joined %v", err, want)
		}
	}
}

func TestWorkloadSupervisesChildFailure(t *testing.T) {
	t.Parallel()
	recovery := newAsyncWorkloadFixture()
	render := newAsyncWorkloadFixture()
	workload, err := NewWorkload(render, recovery)
	if err != nil || workload.Start(context.Background()) != nil {
		t.Fatalf("start workload: %v", err)
	}
	want := errors.New("recovery stopped")
	recovery.failures <- want
	close(render.failures)
	select {
	case got := <-workload.Failures():
		if !errors.Is(got, want) {
			t.Fatalf("failure = %v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("child failure was not supervised")
	}
	_ = workload.Force()
}

func TestWorkloadConcurrentLifecycleUsesStableChildSnapshot(t *testing.T) {
	t.Parallel()
	recovery, render := newAsyncWorkloadFixture(), newAsyncWorkloadFixture()
	recovery.forceErr, render.forceErr = errors.New("force recovery"), errors.New("force render")
	recovery.waitErr, render.waitErr = errors.New("wait recovery"), errors.New("wait render")
	workload, err := NewWorkload(render, recovery)
	if err != nil || workload.Start(context.Background()) != nil {
		t.Fatalf("start workload: %v", err)
	}
	var wait sync.WaitGroup
	forceResults, waitResults := make(chan error, 12), make(chan error, 12)
	for index := 0; index < 12; index++ {
		wait.Add(3)
		go func() { defer wait.Done(); workload.BeginDrain() }()
		go func() { defer wait.Done(); forceResults <- workload.Force() }()
		go func() {
			defer wait.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			waitResults <- workload.Wait(ctx)
		}()
	}
	wait.Wait()
	close(forceResults)
	close(waitResults)
	for result := range forceResults {
		if !errors.Is(result, recovery.forceErr) || !errors.Is(result, render.forceErr) {
			t.Fatalf("concurrent Force() error = %v", result)
		}
	}
	for result := range waitResults {
		if !errors.Is(result, recovery.waitErr) || !errors.Is(result, render.waitErr) {
			t.Fatalf("concurrent Wait() error = %v", result)
		}
	}
	close(recovery.failures)
	close(render.failures)
	if recovery.drainCount() != 12 || render.drainCount() != 12 ||
		recovery.forceCount() != 12 || render.forceCount() != 12 ||
		recovery.waitCount() != 12 || render.waitCount() != 12 {
		t.Fatalf("concurrent lifecycle missed a started child: recovery=%d/%d/%d render=%d/%d/%d",
			recovery.drainCount(), recovery.forceCount(), recovery.waitCount(),
			render.drainCount(), render.forceCount(), render.waitCount())
	}
}

type asyncWorkloadFixture struct {
	startErr        error
	forceErr        error
	waitErr         error
	mutex           sync.Mutex
	started, forced bool
	waited          bool
	drains          int
	forces, waits   int
	failures        chan error
	done            chan struct{}
	doneOnce        sync.Once
}

func newAsyncWorkloadFixture() *asyncWorkloadFixture {
	return &asyncWorkloadFixture{failures: make(chan error, 1), done: make(chan struct{})}
}

func (fixture *asyncWorkloadFixture) Start(context.Context) error {
	if fixture.startErr != nil {
		return fixture.startErr
	}
	fixture.mutex.Lock()
	fixture.started = true
	fixture.mutex.Unlock()
	return nil
}

func (fixture *asyncWorkloadFixture) BeginDrain() {
	fixture.mutex.Lock()
	fixture.drains++
	fixture.mutex.Unlock()
}

func (fixture *asyncWorkloadFixture) Wait(ctx context.Context) error {
	fixture.mutex.Lock()
	fixture.waited = true
	fixture.waits++
	waitErr := fixture.waitErr
	fixture.mutex.Unlock()
	select {
	case <-fixture.done:
		return waitErr
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (fixture *asyncWorkloadFixture) Force() error {
	fixture.mutex.Lock()
	fixture.forced = true
	fixture.forces++
	forceErr := fixture.forceErr
	fixture.mutex.Unlock()
	fixture.doneOnce.Do(func() { close(fixture.done) })
	return forceErr
}

func (fixture *asyncWorkloadFixture) Failures() <-chan error { return fixture.failures }

func (fixture *asyncWorkloadFixture) drainCount() int {
	fixture.mutex.Lock()
	defer fixture.mutex.Unlock()
	return fixture.drains
}

func (fixture *asyncWorkloadFixture) forceCount() int {
	fixture.mutex.Lock()
	defer fixture.mutex.Unlock()
	return fixture.forces
}

func (fixture *asyncWorkloadFixture) waitCount() int {
	fixture.mutex.Lock()
	defer fixture.mutex.Unlock()
	return fixture.waits
}
