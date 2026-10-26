package worker

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type licenseRefreshCycle interface {
	RunOnce(context.Context) (time.Duration, error)
}

type licenseRefreshWorkload struct {
	runner   licenseRefreshCycle
	mutex    sync.Mutex
	cancel   context.CancelFunc
	done     chan struct{}
	failures chan error
	started  bool
}

func newLicenseRefreshWorkload(
	runner licenseRefreshCycle,
) (*licenseRefreshWorkload, error) {
	if runner == nil {
		return nil, fmt.Errorf("license refresh runner is required")
	}
	return &licenseRefreshWorkload{
		runner: runner, done: make(chan struct{}), failures: make(chan error, 1),
	}, nil
}

func (workload *licenseRefreshWorkload) Start(ctx context.Context) error {
	if workload == nil || ctx == nil {
		return fmt.Errorf("license refresh workload is invalid")
	}
	workload.mutex.Lock()
	defer workload.mutex.Unlock()
	if workload.started {
		return fmt.Errorf("license refresh workload is already started")
	}
	runContext, cancel := context.WithCancel(ctx)
	workload.cancel, workload.started = cancel, true
	go workload.run(runContext)
	return nil
}

func (workload *licenseRefreshWorkload) run(ctx context.Context) {
	defer close(workload.done)
	defer close(workload.failures)
	for {
		delay, err := workload.runner.RunOnce(ctx)
		if err != nil {
			select {
			case workload.failures <- err:
			default:
			}
			return
		}
		timer := time.NewTimer(boundedLicenseDelay(delay))
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return
		case <-timer.C:
		}
	}
}

func (workload *licenseRefreshWorkload) BeginDrain() { workload.stop() }

func (workload *licenseRefreshWorkload) Force() error {
	workload.stop()
	return nil
}

func (workload *licenseRefreshWorkload) stop() {
	if workload == nil {
		return
	}
	workload.mutex.Lock()
	cancel := workload.cancel
	workload.mutex.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (workload *licenseRefreshWorkload) Wait(ctx context.Context) error {
	if workload == nil || ctx == nil {
		return fmt.Errorf("license refresh wait is invalid")
	}
	select {
	case <-workload.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (workload *licenseRefreshWorkload) Failures() <-chan error {
	return workload.failures
}

var _ managedAsyncWorkload = (*licenseRefreshWorkload)(nil)
