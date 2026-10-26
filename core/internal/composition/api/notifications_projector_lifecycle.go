package api

import (
	"context"
	"errors"
	"fmt"
)

func (workload *notificationsProjectorWorkload) Start(parent context.Context) error {
	if parent == nil {
		return fmt.Errorf("notification projector context is required")
	}
	workload.mutex.Lock()
	defer workload.mutex.Unlock()
	if workload.started {
		return fmt.Errorf("notification projector is already started")
	}
	ctx, cancel := context.WithCancel(parent)
	workload.cancel = cancel
	workload.done = make(chan struct{})
	workload.started = true
	go workload.runManaged(ctx)
	return nil
}

func (workload *notificationsProjectorWorkload) runManaged(ctx context.Context) {
	defer close(workload.done)
	defer close(workload.failures)
	if err := workload.run(ctx); err != nil {
		workload.failures <- err
	}
}

func (workload *notificationsProjectorWorkload) Failures() <-chan error {
	return workload.failures
}

func (workload *notificationsProjectorWorkload) BeginDrain() {
	workload.realtime.BeginDrain()
	workload.mutex.Lock()
	defer workload.mutex.Unlock()
	if workload.cancel != nil {
		workload.cancel()
	}
}

func (workload *notificationsProjectorWorkload) Wait(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("notification projector wait context is required")
	}
	workload.mutex.Lock()
	done := workload.done
	workload.mutex.Unlock()
	var projectorErr error
	if done != nil {
		select {
		case <-done:
		case <-ctx.Done():
			projectorErr = ctx.Err()
		}
	}
	return errors.Join(projectorErr, workload.realtime.Wait(ctx))
}

func (workload *notificationsProjectorWorkload) Force() error {
	workload.mutex.Lock()
	if workload.cancel != nil {
		workload.cancel()
	}
	workload.mutex.Unlock()
	return workload.realtime.Force()
}
