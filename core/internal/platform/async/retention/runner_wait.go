package retention

import "context"

func (runner *Runner) Wait(ctx context.Context) error {
	if ctx == nil {
		return ErrDependencies
	}
	runner.lifecycle.mutex.Lock()
	done := runner.lifecycle.done
	runner.lifecycle.mutex.Unlock()
	if done == nil {
		return nil
	}
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (runner *Runner) Force() error {
	runner.lifecycle.mutex.Lock()
	defer runner.lifecycle.mutex.Unlock()
	if runner.lifecycle.cancel != nil {
		runner.lifecycle.cancel()
	}
	return nil
}
