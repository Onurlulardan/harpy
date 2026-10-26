package execution

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRegistryRejectsAsyncProgramRunAndCancelsCapturedContext(t *testing.T) {
	started := make(chan struct{})
	runDone := make(chan struct{})
	attempt := make(chan error, 1)
	release := make(chan struct{})
	rowsClosed := make(chan struct{})
	program := newTestProgram(t, "forms", func(
		ctx context.Context,
		driver SnapshotDriver,
	) error {
		if _, err := driver.Query(ctx, "SELECT 1"); err != nil {
			return err
		}
		close(started)
		<-ctx.Done()
		attempt <- driver.Execute(ctx, "SELECT 1")
		<-release
		return ctx.Err()
	}, testScopeVerifier)
	driver := &testDriver{
		owner: "forms", rows: &testRows{closedSignal: rowsClosed},
	}
	registry := newTestRegistry(t, "forms", testPort(func(
		_ transaction.TenantScope,
		received Program,
	) error {
		go func() {
			defer close(runDone)
			_ = received.Run(driver)
		}()
		<-started
		return nil
	}))
	if err := registry.Run(program); !errors.Is(err, errExecutionContract) {
		t.Fatalf("async Run error=%v", err)
	}
	select {
	case err := <-attempt:
		if !errors.Is(err, errExecutionContract) {
			t.Fatalf("canceled driver remained usable: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("async closure did not observe cancellation")
	}
	select {
	case <-rowsClosed:
	case <-time.After(time.Second):
		t.Fatal("cancellation did not close leased rows")
	}
	close(release)
	select {
	case <-runDone:
	case <-time.After(time.Second):
		t.Fatal("async closure did not stop")
	}
}

func TestProgramConcurrentDispatchHasOneWinner(t *testing.T) {
	var executions atomic.Int32
	program := newTestProgram(t, "forms", func(
		context.Context, SnapshotDriver,
	) error {
		executions.Add(1)
		return nil
	}, testScopeVerifier)
	var portCalls atomic.Int32
	registry := newTestRegistry(t, "forms", testPort(func(
		_ transaction.TenantScope,
		received Program,
	) error {
		portCalls.Add(1)
		return received.Run(&testDriver{owner: "forms"})
	}))
	const callers = 32
	var successes atomic.Int32
	var wait sync.WaitGroup
	wait.Add(callers)
	for range callers {
		go func() {
			defer wait.Done()
			if registry.Run(program) == nil {
				successes.Add(1)
			}
		}()
	}
	wait.Wait()
	if successes.Load() != 1 || portCalls.Load() != 1 || executions.Load() != 1 {
		t.Fatalf(
			"success=%d port=%d execute=%d",
			successes.Load(), portCalls.Load(), executions.Load(),
		)
	}
}

func TestProgramPanicCompletesReceiptAndCancelsContext(t *testing.T) {
	program := newTestProgram(t, "forms", func(
		context.Context, SnapshotDriver,
	) error {
		panic("executor panic")
	}, testScopeVerifier)
	registry := newTestRegistry(t, "forms", testPort(func(
		_ transaction.TenantScope,
		received Program,
	) error {
		return received.Run(&testDriver{owner: "forms"})
	}))
	func() {
		defer func() {
			if recover() != "executor panic" {
				t.Fatal("executor panic was not preserved")
			}
		}()
		_ = registry.Run(program)
	}()
	if program.state.context.Err() == nil ||
		programPhase(program.state.phase.Load()) != programCompleted {
		t.Fatal("panic did not close Program lifetime")
	}
	if err := registry.Run(program); !errors.Is(err, errExecutionContract) {
		t.Fatalf("panic replay error=%v", err)
	}
}
