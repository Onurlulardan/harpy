package execution

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestProgramStageRejectsDuplicateAndMixedSnapshotPrograms(t *testing.T) {
	if _, err := NewProgramStage([]Program{{}}); !errors.Is(
		err, errExecutionContract,
	) {
		t.Fatalf("zero program error=%v", err)
	}
	one := newTestProgram(t, "forms", func(context.Context, SnapshotDriver) error {
		return nil
	}, testScopeVerifier)
	if _, err := NewProgramStage([]Program{one, one}); !errors.Is(
		err, errExecutionContract,
	) {
		t.Fatalf("duplicate error=%v", err)
	}
	two := newTestProgram(t, "crm", func(context.Context, SnapshotDriver) error {
		return nil
	}, testScopeVerifier)
	two.state.identity.operationID = "restquery.other"
	if _, err := NewProgramStage([]Program{one, two}); !errors.Is(
		err, errExecutionContract,
	) {
		t.Fatalf("mixed snapshot error=%v", err)
	}
}

func TestRegistryRunStagePreflightsAllPortsBeforeCallbacks(t *testing.T) {
	var calls atomic.Int32
	forms := newTestProgram(t, "forms", func(context.Context, SnapshotDriver) error {
		calls.Add(1)
		return nil
	}, testScopeVerifier)
	crm := newTestProgram(t, "crm", func(context.Context, SnapshotDriver) error {
		calls.Add(1)
		return nil
	}, testScopeVerifier)
	stage, err := NewProgramStage([]Program{forms, crm})
	if err != nil {
		t.Fatal(err)
	}
	registry := newTestRegistry(t, "forms", testPort(func(
		_ transaction.TenantScope, program Program,
	) error {
		return program.Run(&testDriver{owner: "forms"})
	}))
	if err := registry.RunStage(stage); !errors.Is(err, errExecutionContract) {
		t.Fatalf("missing port error=%v", err)
	}
	if calls.Load() != 0 || forms.state.context.Err() == nil ||
		crm.state.context.Err() == nil {
		t.Fatalf("calls=%d forms=%v crm=%v", calls.Load(),
			forms.state.context.Err(), crm.state.context.Err())
	}
}

func TestRegistryRunStageDispatchesCallbacksConcurrently(t *testing.T) {
	started := make(chan string, 2)
	release := make(chan struct{})
	callback := func(owner string) func(context.Context, SnapshotDriver) error {
		return func(context.Context, SnapshotDriver) error {
			started <- owner
			<-release
			return nil
		}
	}
	forms := newTestProgram(t, "forms", callback("forms"), testScopeVerifier)
	crm := newTestProgram(t, "crm", callback("crm"), testScopeVerifier)
	stage, err := NewProgramStage([]Program{forms, crm})
	if err != nil {
		t.Fatal(err)
	}
	registry := stageTestRegistry(t)
	done := make(chan error, 1)
	go func() { done <- registry.RunStage(stage) }()
	seen := map[string]bool{<-started: true, <-started: true}
	if !seen["forms"] || !seen["crm"] {
		t.Fatalf("started=%v", seen)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func stageTestRegistry(t *testing.T) Registry {
	t.Helper()
	registrations := make([]Registration, 0, 2)
	for _, owner := range []string{"forms", "crm"} {
		current := owner
		registration, err := NewRegistration(current, testPort(func(
			_ transaction.TenantScope, program Program,
		) error {
			return program.Run(&testDriver{owner: current})
		}))
		if err != nil {
			t.Fatal(err)
		}
		registrations = append(registrations, registration)
	}
	registry, err := NewRegistry(registrations)
	if err != nil {
		t.Fatal(err)
	}
	return registry
}
