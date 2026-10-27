package eventdelivery

import (
	"context"
	"errors"
)

func (runner *Runner) Wait(ctx context.Context) error {
	if ctx == nil {
		return ErrRunnerDependencies
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
	sessions := make([]brokerSession, 0, len(runner.lifecycle.active))
	for key, session := range runner.lifecycle.active {
		sessions = append(sessions, session)
		delete(runner.lifecycle.active, key)
	}
	runner.lifecycle.mutex.Unlock()
	var result error
	for _, session := range sessions {
		result = errors.Join(result, session.Close())
	}
	return result
}
