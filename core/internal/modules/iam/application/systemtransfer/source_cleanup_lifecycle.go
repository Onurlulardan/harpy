package systemtransfer

import "context"

func (runner *SourceCleanupRunner) Failures() <-chan error {
	return runner.failures
}

func (runner *SourceCleanupRunner) BeginDrain() {
	if runner == nil {
		return
	}
	runner.mutex.Lock()
	defer runner.mutex.Unlock()
	if runner.scanCancel != nil {
		runner.scanCancel()
	}
}

func (runner *SourceCleanupRunner) Force() error {
	if runner == nil {
		return nil
	}
	runner.mutex.Lock()
	defer runner.mutex.Unlock()
	if runner.scanCancel != nil {
		runner.scanCancel()
	}
	if runner.forceCancel != nil {
		runner.forceCancel()
	}
	return nil
}

func (runner *SourceCleanupRunner) Wait(ctx context.Context) error {
	if runner == nil || ctx == nil {
		return ErrSourceCleanupInvalid
	}
	select {
	case <-runner.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
