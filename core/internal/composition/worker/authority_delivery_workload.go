package worker

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type authorityDeliveryCycle interface {
	RunOnce(context.Context) (time.Duration, error)
}

type authorityDeliveryWorkload struct {
	runner   authorityDeliveryCycle
	mutex    sync.Mutex
	cancel   context.CancelFunc
	done     chan struct{}
	failures chan error
	started  bool
}

func newAuthorityDeliveryWorkload(
	runner authorityDeliveryCycle,
) (*authorityDeliveryWorkload, error) {
	if runner == nil {
		return nil, fmt.Errorf("platform authority delivery runner is required")
	}
	return &authorityDeliveryWorkload{
		runner: runner, done: make(chan struct{}), failures: make(chan error, 1),
	}, nil
}

func (workload *authorityDeliveryWorkload) Start(ctx context.Context) error {
	if workload == nil || ctx == nil {
		return fmt.Errorf("platform authority delivery workload is invalid")
	}
	workload.mutex.Lock()
	defer workload.mutex.Unlock()
	if workload.started {
		return fmt.Errorf("platform authority delivery workload is already started")
	}
	runContext, cancel := context.WithCancel(ctx)
	workload.cancel, workload.started = cancel, true
	go workload.run(runContext)
	return nil
}

func (workload *authorityDeliveryWorkload) run(ctx context.Context) {
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

func (workload *authorityDeliveryWorkload) BeginDrain() { workload.stop() }
func (workload *authorityDeliveryWorkload) Force() error {
	workload.stop()
	return nil
}

func (workload *authorityDeliveryWorkload) stop() {
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

func (workload *authorityDeliveryWorkload) Wait(ctx context.Context) error {
	if workload == nil || ctx == nil {
		return fmt.Errorf("platform authority delivery wait is invalid")
	}
	select {
	case <-workload.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (workload *authorityDeliveryWorkload) Failures() <-chan error {
	return workload.failures
}

var _ managedAsyncWorkload = (*authorityDeliveryWorkload)(nil)
