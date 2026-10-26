package outboxdelivery

import (
	"context"
	"errors"
	"time"
)

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
	if runner.lifecycle.cancel != nil {
		runner.lifecycle.cancel()
	}
	active := runner.lifecycle.active
	runner.lifecycle.active = nil
	runner.lifecycle.mutex.Unlock()
	if active == nil {
		return nil
	}
	return active.Close()
}

func waitContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func cleanStop(err error, ctx context.Context) bool {
	return errors.Is(err, context.Canceled) && ctx.Err() != nil
}
