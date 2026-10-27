package deadletterdelivery

import (
	"context"
	"sync"
)

type runnerLifecycle struct {
	mutex    *sync.Mutex
	cancel   context.CancelFunc
	done     chan struct{}
	failures chan error
	active   Session
	started  bool
}

func newRunnerLifecycle() runnerLifecycle {
	return runnerLifecycle{
		mutex:    &sync.Mutex{},
		failures: make(chan error, 1),
	}
}

func (runner *Runner) Start(parent context.Context) error {
	if parent == nil {
		return ErrDependencies
	}
	runner.lifecycle.mutex.Lock()
	defer runner.lifecycle.mutex.Unlock()
	if runner.lifecycle.started {
		return ErrAlreadyStarted
	}
	ctx, cancel := context.WithCancel(parent)
	runner.lifecycle.cancel = cancel
	runner.lifecycle.done = make(chan struct{})
	runner.lifecycle.started = true
	go runner.runManaged(ctx)
	return nil
}

func (runner *Runner) runManaged(ctx context.Context) {
	defer close(runner.lifecycle.done)
	defer close(runner.lifecycle.failures)
	defer runner.recoverManagedPanic()
	if err := runner.run(ctx); err != nil && ctx.Err() == nil {
		runner.lifecycle.failures <- err
	}
}

func (runner *Runner) recoverManagedPanic() {
	if recover() == nil {
		return
	}
	active := runner.detachActive()
	_ = closeSession(active)
	runner.lifecycle.failures <- ErrSessionPanicked
}

func (runner *Runner) Failures() <-chan error {
	return runner.lifecycle.failures
}

func (runner *Runner) BeginDrain() {
	runner.lifecycle.mutex.Lock()
	if runner.lifecycle.cancel != nil {
		runner.lifecycle.cancel()
	}
	active := runner.lifecycle.active
	runner.lifecycle.active = nil
	runner.lifecycle.mutex.Unlock()
	if active != nil {
		_ = closeSession(active)
	}
}
