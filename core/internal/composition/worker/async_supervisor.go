package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type managedAsyncWorkload interface {
	Start(context.Context) error
	BeginDrain()
	Wait(context.Context) error
	Force() error
	Failures() <-chan error
}

type asyncSupervisor struct {
	workloads []managedAsyncWorkload
	failures  chan error
	started   bool
	mutex     *sync.Mutex
}

func newAsyncSupervisor(workloads ...managedAsyncWorkload) (*asyncSupervisor, error) {
	if len(workloads) == 0 {
		return nil, fmt.Errorf("async supervisor requires workloads")
	}
	for _, workload := range workloads {
		if workload == nil {
			return nil, fmt.Errorf("async supervisor workload is nil")
		}
	}
	return &asyncSupervisor{
		workloads: append([]managedAsyncWorkload(nil), workloads...),
		failures:  make(chan error, 1), mutex: &sync.Mutex{},
	}, nil
}

func (supervisor *asyncSupervisor) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("async supervisor context is nil")
	}
	supervisor.mutex.Lock()
	defer supervisor.mutex.Unlock()
	if supervisor.started {
		return fmt.Errorf("async supervisor is already started")
	}
	started := make([]managedAsyncWorkload, 0, len(supervisor.workloads))
	for _, workload := range supervisor.workloads {
		if err := workload.Start(ctx); err != nil {
			for _, active := range started {
				_ = active.Force()
			}
			return fmt.Errorf("start async workload: %w", err)
		}
		started = append(started, workload)
	}
	supervisor.started = true
	go supervisor.monitorFailures()
	return nil
}

func (supervisor *asyncSupervisor) monitorFailures() {
	var wait sync.WaitGroup
	for _, workload := range supervisor.workloads {
		wait.Add(1)
		go func(active managedAsyncWorkload) {
			defer wait.Done()
			if err, open := <-active.Failures(); open && err != nil {
				select {
				case supervisor.failures <- err:
				default:
				}
			}
		}(workload)
	}
	wait.Wait()
	close(supervisor.failures)
}

func (supervisor *asyncSupervisor) Failures() <-chan error { return supervisor.failures }

func (supervisor *asyncSupervisor) BeginDrain() {
	for _, workload := range supervisor.workloads {
		workload.BeginDrain()
	}
}

func (supervisor *asyncSupervisor) Wait(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("async supervisor wait context is nil")
	}
	errorsByWorkload := make(chan error, len(supervisor.workloads))
	for _, workload := range supervisor.workloads {
		go func(active managedAsyncWorkload) { errorsByWorkload <- active.Wait(ctx) }(workload)
	}
	var result error
	for range supervisor.workloads {
		result = errors.Join(result, <-errorsByWorkload)
	}
	return result
}

func (supervisor *asyncSupervisor) Force() error {
	var result error
	for _, workload := range supervisor.workloads {
		result = errors.Join(result, workload.Force())
	}
	return result
}
