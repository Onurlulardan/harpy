package worker

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type customerProvisioningRunner interface {
	RunOnce(context.Context) (bool, error)
}

type customerProvisioningWorkload struct {
	runner   customerProvisioningRunner
	idle     time.Duration
	mutex    sync.Mutex
	cancel   context.CancelFunc
	done     chan struct{}
	failures chan error
	started  bool
}

func newCustomerProvisioningWorkload(
	runner customerProvisioningRunner,
) (*customerProvisioningWorkload, error) {
	if runner == nil {
		return nil, fmt.Errorf("customer provisioning runner is required")
	}
	return &customerProvisioningWorkload{
		runner: runner, idle: time.Second,
		done: make(chan struct{}), failures: make(chan error, 1),
	}, nil
}

func (workload *customerProvisioningWorkload) Start(ctx context.Context) error {
	if workload == nil || ctx == nil {
		return fmt.Errorf("customer provisioning workload is invalid")
	}
	workload.mutex.Lock()
	defer workload.mutex.Unlock()
	if workload.started {
		return fmt.Errorf("customer provisioning workload is already started")
	}
	runContext, cancel := context.WithCancel(ctx)
	workload.cancel = cancel
	workload.started = true
	go workload.run(runContext)
	return nil
}

func (workload *customerProvisioningWorkload) run(ctx context.Context) {
	defer close(workload.done)
	defer close(workload.failures)
	for {
		worked, err := workload.runner.RunOnce(ctx)
		if err != nil {
			select {
			case workload.failures <- err:
			default:
			}
			return
		}
		delay := workload.idle
		if worked {
			delay = 10 * time.Millisecond
		}
		timer := time.NewTimer(delay)
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

func (workload *customerProvisioningWorkload) BeginDrain() { workload.stop() }

func (workload *customerProvisioningWorkload) Force() error {
	workload.stop()
	return nil
}

func (workload *customerProvisioningWorkload) stop() {
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

func (workload *customerProvisioningWorkload) Wait(ctx context.Context) error {
	if workload == nil || ctx == nil {
		return fmt.Errorf("customer provisioning wait is invalid")
	}
	select {
	case <-workload.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (workload *customerProvisioningWorkload) Failures() <-chan error {
	return workload.failures
}

var _ managedAsyncWorkload = (*customerProvisioningWorkload)(nil)
