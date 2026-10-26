package jobexecution

import (
	"context"
	"sync"
)

type runnerLifecycle struct {
	mutex       *sync.Mutex
	drainCancel context.CancelFunc
	forceCancel context.CancelFunc
	done        chan struct{}
	failures    chan error
	started     bool
}

func newRunnerLifecycle() runnerLifecycle {
	return runnerLifecycle{mutex: &sync.Mutex{}, failures: make(chan error, 1)}
}

func (runner *Runner) Start(parent context.Context) error {
	if parent == nil {
		return ErrRunnerDependencies
	}
	runner.lifecycle.mutex.Lock()
	defer runner.lifecycle.mutex.Unlock()
	if runner.lifecycle.started {
		return ErrRunnerStarted
	}
	forceContext, forceCancel := context.WithCancel(parent)
	drainContext, drainCancel := context.WithCancel(forceContext)
	runner.lifecycle.forceCancel = forceCancel
	runner.lifecycle.drainCancel = drainCancel
	runner.lifecycle.done = make(chan struct{})
	runner.lifecycle.started = true
	go runner.runManaged(drainContext, forceContext, forceCancel)
	return nil
}

func (runner *Runner) runManaged(
	drainContext context.Context,
	forceContext context.Context,
	forceCancel context.CancelFunc,
) {
	defer close(runner.lifecycle.done)
	defer close(runner.lifecycle.failures)
	if err := runner.run(drainContext, forceContext, forceCancel); err != nil {
		runner.lifecycle.failures <- err
	}
}

func (runner *Runner) Failures() <-chan error { return runner.lifecycle.failures }

func (runner *Runner) BeginDrain() {
	runner.lifecycle.mutex.Lock()
	defer runner.lifecycle.mutex.Unlock()
	if runner.lifecycle.drainCancel != nil {
		runner.lifecycle.drainCancel()
	}
}
