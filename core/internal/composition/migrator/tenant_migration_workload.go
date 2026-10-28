package migrator

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	"github.com/Onurlulardan/harpy/core/internal/platform/processhost"
	tenantexec "github.com/Onurlulardan/harpy/core/internal/platform/tenantmigrationexecutor"
)

type tenantMigrationLoop interface {
	Run(context.Context, tenantexec.Admission) error
}

type tenantMigrationWorkload struct {
	gate       *admission.Gate
	cancel     context.CancelFunc
	done       chan struct{}
	failures   chan error
	mutex      sync.Mutex
	drainError error
}

func startTenantMigrationWorkload(
	loop tenantMigrationLoop, gate *admission.Gate,
) (*tenantMigrationWorkload, error) {
	if loop == nil || gate == nil {
		return nil, fmt.Errorf("tenant migration workload dependencies are incomplete")
	}
	ctx, cancel := context.WithCancel(context.Background())
	workload := &tenantMigrationWorkload{gate: gate, cancel: cancel,
		done: make(chan struct{}), failures: make(chan error, 1)}
	go workload.run(ctx, loop)
	return workload, nil
}

func (workload *tenantMigrationWorkload) run(
	ctx context.Context, loop tenantMigrationLoop,
) {
	defer close(workload.done)
	defer close(workload.failures)
	err := loop.Run(ctx, workload.gate)
	if err == nil && ctx.Err() == nil && !workload.gate.Draining() {
		err = fmt.Errorf("tenant migration executor stopped before drain")
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
	agentDebugLog("I", "tenant_migration_workload.go:run", "executor stopped", map[string]string{
		"error": err.Error(),
	})
	workload.failures <- err
}

func (workload *tenantMigrationWorkload) lifecycle() processhost.Workload {
	return processhost.Workload{BeginDrain: workload.gate.BeginDrain,
		Wait: workload.wait, Force: func() error { workload.cancel(); return nil },
		Failures: workload.failures}
}

func (workload *tenantMigrationWorkload) wait(ctx context.Context) error {
	select {
	case <-workload.done:
		workload.mutex.Lock()
		defer workload.mutex.Unlock()
		return workload.drainError
	case <-ctx.Done():
		return ctx.Err()
	}
}
