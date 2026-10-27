package systemtransfer

import (
	"context"
	"fmt"
)

func NewSourceCleanupRunner(store SourceCleanupStore,
	objects SourceObjectDeleter,
	clock SourceCleanupClock,
	ids SourceCleanupIDs,
	jitter SourceCleanupJitter,
	gate SourceCleanupAdmission,
) (*SourceCleanupRunner, error) {
	for _, dependency := range []any{store, objects, clock, ids, jitter, gate} {
		if nilDependency(dependency) {
			return nil, ErrSourceCleanupUnavailable
		}
	}
	ownerID, err := ids.NewV7(clock.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("create IAM import source cleanup owner: %w", err)
	}
	if ownerID.Version() != 7 {
		return nil, ErrSourceCleanupInvalid
	}
	return &SourceCleanupRunner{store: store, objects: objects, clock: clock,
		jitter: jitter, ownerID: ownerID, admission: gate, done: make(chan struct{}),
		failures: make(chan error, 1)}, nil
}

func (runner *SourceCleanupRunner) Start(parent context.Context) error {
	if runner == nil || parent == nil || parent.Err() != nil {
		return ErrSourceCleanupInvalid
	}
	runner.mutex.Lock()
	defer runner.mutex.Unlock()
	if runner.started {
		return ErrSourceCleanupInvalid
	}
	scan, scanCancel := context.WithCancel(parent)
	force, forceCancel := context.WithCancel(parent)
	runner.scanCancel, runner.forceCancel = scanCancel, forceCancel
	runner.started = true
	go runner.serve(scan, force)
	return nil
}

func (runner *SourceCleanupRunner) serve(scan, force context.Context) {
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
