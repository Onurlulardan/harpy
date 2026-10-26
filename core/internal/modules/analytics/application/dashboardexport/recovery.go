package dashboardexport

import (
	"context"
	"fmt"
	"sync"
)

type RecoveryRunner struct {
	clock       Clock
	ids         IDSource
	store       RecoveryStore
	artifacts   ArtifactRuntime
	spools      SpoolAllocator
	admission   DispatcherAdmission
	mutex       sync.Mutex
	started     bool
	scanCancel  context.CancelFunc
	forceCancel context.CancelFunc
	done        chan struct{}
	failures    chan error
}

func NewRecoveryRunner(
	clock Clock, ids IDSource, store RecoveryStore, artifacts ArtifactRuntime,
	spools SpoolAllocator, admission DispatcherAdmission,
) (*RecoveryRunner, error) {
	for _, dependency := range []any{clock, ids, store, artifacts, spools, admission} {
		if nilPort(dependency) {
			return nil, fmt.Errorf("Analytics dashboard export recovery dependency is unavailable")
		}
	}
	return &RecoveryRunner{clock: clock, ids: ids, store: store,
		artifacts: artifacts, spools: spools, admission: admission,
		done: make(chan struct{}), failures: make(chan error, 1)}, nil
}

func (runner *RecoveryRunner) Start(parent context.Context) error {
	if parent == nil || parent.Err() != nil {
		return ErrInvalid
	}
	runner.mutex.Lock()
	defer runner.mutex.Unlock()
	if runner.started {
		return fmt.Errorf("Analytics dashboard export recovery is already started")
	}
	scan, scanCancel := context.WithCancel(parent)
	force, forceCancel := context.WithCancel(parent)
	runner.scanCancel, runner.forceCancel = scanCancel, forceCancel
	runner.started = true
	go runner.serve(scan, force)
	return nil
}

func (runner *RecoveryRunner) serve(scan, force context.Context) {
	err := runner.run(scan, force)
	if err != nil {
		select {
		case runner.failures <- err:
		default:
		}
	}
	close(runner.failures)
	close(runner.done)
}

func (runner *RecoveryRunner) Failures() <-chan error { return runner.failures }

func (runner *RecoveryRunner) BeginDrain() {
	runner.mutex.Lock()
	defer runner.mutex.Unlock()
	if runner.scanCancel != nil {
		runner.scanCancel()
	}
}

func (runner *RecoveryRunner) Force() error {
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

func (runner *RecoveryRunner) Wait(ctx context.Context) error {
	if ctx == nil {
		return ErrInvalid
	}
	select {
	case <-runner.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
