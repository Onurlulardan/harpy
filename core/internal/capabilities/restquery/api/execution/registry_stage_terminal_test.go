package execution

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRegistryRunStagePrioritizesSkippedRunAtEveryIndex(t *testing.T) {
	owners := []string{"forms", "crm", "workflow", "documents"}
	for skipped := range owners {
		t.Run(owners[skipped], func(t *testing.T) {
			stage, registry := terminalStageFixture(t, owners, skipped, -1)
			if err := registry.RunStage(stage); !errors.Is(err, errExecutionContract) {
				t.Fatalf("terminal error=%v", err)
			}
		})
	}
}

func TestRegistryRunStagePreservesHighIndexPanicOverLowIndexCancel(t *testing.T) {
	owners := []string{"forms", "crm"}
	stage, registry := terminalStageFixture(t, owners, -1, 1)
	panicValue := make(chan any, 1)
	go func() {
		defer func() { panicValue <- recover() }()
		_ = registry.RunStage(stage)
	}()
	select {
	case value := <-panicValue:
		if value != "terminal panic" {
			t.Fatalf("panic=%v", value)
		}
	case <-time.After(time.Second):
		t.Fatal("terminal panic cleanup deadlocked")
	}
}

func terminalStageFixture(
	t *testing.T,
	owners []string,
	skipped int,
	panicIndex int,
) (ProgramStage, Registry) {
	t.Helper()
	programs := make([]Program, len(owners))
	registrations := make([]Registration, len(owners))
	for index, owner := range owners {
		currentIndex, currentOwner := index, owner
		programs[index] = newTestProgram(t, currentOwner, func(
			ctx context.Context,
			_ SnapshotDriver,
		) error {
			if currentIndex == panicIndex {
				panic("terminal panic")
			}
			<-ctx.Done()
			return ctx.Err()
		}, testScopeVerifier)
		registration, err := NewRegistration(currentOwner, testPort(func(
			_ transaction.TenantScope,
			program Program,
		) error {
			if currentIndex == skipped {
				return nil
			}
			return program.Run(&testDriver{owner: currentOwner})
		}))
		if err != nil {
			t.Fatal(err)
		}
		registrations[index] = registration
	}
	stage, err := NewProgramStage(programs)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := NewRegistry(registrations)
	if err != nil {
		t.Fatal(err)
	}
	return stage, registry
}
