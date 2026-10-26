package retention

import "context"

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
	if err := runner.run(ctx); err != nil {
		runner.lifecycle.failures <- err
	}
}

func (runner *Runner) Failures() <-chan error { return runner.lifecycle.failures }

func (runner *Runner) BeginDrain() {
	runner.lifecycle.mutex.Lock()
	defer runner.lifecycle.mutex.Unlock()
	if runner.lifecycle.cancel != nil {
		runner.lifecycle.cancel()
	}
}
