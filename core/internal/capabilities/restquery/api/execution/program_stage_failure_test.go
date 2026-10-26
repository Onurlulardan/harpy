package execution

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRegistryRunStageCancelsSiblingWhenPortSkipsProgramRun(t *testing.T) {
	forms := newTestProgram(t, "forms", func(
		ctx context.Context, _ SnapshotDriver,
	) error {
		<-ctx.Done()
		return ctx.Err()
	}, testScopeVerifier)
	crm := newTestProgram(t, "crm", func(context.Context, SnapshotDriver) error {
		return nil
	}, testScopeVerifier)
	stage, err := NewProgramStage([]Program{forms, crm})
	if err != nil {
		t.Fatal(err)
	}
	registry := stageFailureRegistry(t, false)
	done := make(chan error, 1)
	go func() { done <- registry.RunStage(stage) }()
	select {
	case err := <-done:
		if !errors.Is(err, errExecutionContract) {
			t.Fatalf("error=%v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("missing Program.Run deadlocked sibling")
	}
	if forms.state.context.Err() == nil || crm.state.context.Err() == nil {
		t.Fatal("stage contexts were not canceled")
	}
}

func TestRegistryRunStagePreservesPanicAfterSiblingCleanup(t *testing.T) {
	panicProgram := newTestProgram(t, "forms", func(
		context.Context, SnapshotDriver,
	) error {
		panic("stage panic")
	}, testScopeVerifier)
	waitProgram := newTestProgram(t, "crm", func(
		ctx context.Context, _ SnapshotDriver,
	) error {
		<-ctx.Done()
		return ctx.Err()
	}, testScopeVerifier)
	stage, err := NewProgramStage([]Program{panicProgram, waitProgram})
	if err != nil {
		t.Fatal(err)
	}
	registry := stageFailureRegistry(t, true)
	panicValue := make(chan any, 1)
	go func() {
		defer func() { panicValue <- recover() }()
		_ = registry.RunStage(stage)
	}()
	select {
	case value := <-panicValue:
		if value != "stage panic" || waitProgram.state.context.Err() == nil {
			t.Fatalf("panic=%v sibling=%v", value, waitProgram.state.context.Err())
		}
	case <-time.After(time.Second):
		t.Fatal("panic cleanup deadlocked")
	}
}

func stageFailureRegistry(t *testing.T, runCRM bool) Registry {
	t.Helper()
	forms, _ := NewRegistration("forms", testPort(func(
		_ transaction.TenantScope, program Program,
	) error {
		return program.Run(&testDriver{owner: "forms"})
	}))
	crm, _ := NewRegistration("crm", testPort(func(
		_ transaction.TenantScope, program Program,
	) error {
		if !runCRM {
			return nil
		}
		return program.Run(&testDriver{owner: "crm"})
	}))
	registry, err := NewRegistry([]Registration{forms, crm})
	if err != nil {
		t.Fatal(err)
	}
	return registry
}
