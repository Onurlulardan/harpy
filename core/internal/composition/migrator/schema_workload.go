package migrator

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	"github.com/Onurlulardan/harpy/core/internal/platform/processhost"
)

type schemaLoop interface {
	Run(context.Context, *admission.Gate) error
}

type schemaWorkload struct {
	gate       *admission.Gate
	cancel     context.CancelFunc
	done       chan struct{}
	failures   chan error
	mutex      sync.Mutex
	drainError error
}

func startSchemaWorkload(loop schemaLoop, gate *admission.Gate) (*schemaWorkload, error) {
	if loop == nil || gate == nil {
		return nil, fmt.Errorf("schema workload dependencies are incomplete")
	}
	workContext, cancel := context.WithCancel(context.Background())
	workload := &schemaWorkload{
		gate: gate, cancel: cancel, done: make(chan struct{}), failures: make(chan error, 1),
	}
	go workload.run(workContext, loop)
	return workload, nil
}

func (workload *schemaWorkload) run(ctx context.Context, loop schemaLoop) {
	defer close(workload.done)
	defer close(workload.failures)
	err := loop.Run(ctx, workload.gate)
	if err == nil && ctx.Err() == nil && !workload.gate.Draining() {
		err = fmt.Errorf("schema executor stopped before drain")
	}
	if err == nil || errors.Is(err, context.Canceled) {
		return
	}
	if workload.gate.Draining() {
		workload.mutex.Lock()
		workload.drainError = err
		workload.mutex.Unlock()
		return
	}
	workload.failures <- err
}

func (workload *schemaWorkload) lifecycle() processhost.Workload {
	return processhost.Workload{
		BeginDrain: workload.gate.BeginDrain,
		Wait:       workload.wait,
		Force: func() error {
			workload.cancel()
			return nil
		},
		Failures: workload.failures,
	}
}

func (workload *schemaWorkload) wait(ctx context.Context) error {
	select {
	case <-workload.done:
		workload.mutex.Lock()
		defer workload.mutex.Unlock()
		return workload.drainError
	case <-ctx.Done():
		return ctx.Err()
	}
}

type idleSchemaLoop struct{}

func (idleSchemaLoop) Run(ctx context.Context, gate *admission.Gate) error {
	if gate == nil {
		return fmt.Errorf("schema executor admission is unavailable")
	}
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		if gate.Draining() {
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
