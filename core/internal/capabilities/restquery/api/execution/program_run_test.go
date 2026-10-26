package execution

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRegistryRunsProgramOnceWithCapturedScopeAndContext(t *testing.T) {
	driver := &testDriver{owner: "forms"}
	var escapedDriver SnapshotDriver
	var escapedRows Rows
	var escapedRow Row
	program := newTestProgram(t, "forms", func(
		ctx context.Context,
		leased SnapshotDriver,
	) error {
		if ctx.Value(testContextKey{}) != "captured" {
			t.Fatal("captured context was replaced")
		}
		if err := leased.Execute(context.Background(), "SET LOCAL x = y"); err != nil {
			return err
		}
		rows, err := leased.Query(context.Background(), "SELECT 1")
		if err != nil {
			return err
		}
		row := leased.QueryRow(context.Background(), "SELECT 1")
		escapedDriver, escapedRows, escapedRow = leased, rows, row
		return nil
	}, testScopeVerifier)
	portCalls := 0
	registry := newTestRegistry(t, "forms", testPort(func(
		scope transaction.TenantScope,
		received Program,
	) error {
		portCalls++
		if scope != testScopeIdentity().scope {
			t.Fatal("captured scope was replaced")
		}
		return received.Run(driver)
	}))
	copyOfProgram := program
	if err := registry.Run(program); err != nil {
		t.Fatal(err)
	}
	if portCalls != 1 || driver.executeCalls.Load() != 1 ||
		driver.context.Value(testContextKey{}) != "captured" {
		t.Fatal("owner port did not use captured execution state")
	}
	if err := registry.Run(copyOfProgram); !errors.Is(err, errExecutionContract) {
		t.Fatalf("replay error=%v", err)
	}
	if err := escapedDriver.Execute(context.Background(), "SELECT 1"); !errors.Is(err, errExecutionContract) {
		t.Fatalf("escaped driver error=%v", err)
	}
	if escapedRows.Next() || !errors.Is(escapedRows.Err(), errExecutionContract) ||
		!errors.Is(escapedRow.Scan(), errExecutionContract) || !driver.rows.closed.Load() {
		t.Fatal("driver or row escaped Program lifetime")
	}
}

func TestProgramOutcomeCannotBeSwallowedByOwnerPort(t *testing.T) {
	wanted := errors.New("typed query failure")
	driver := &testDriver{owner: "forms"}
	program := newTestProgram(t, "forms", func(
		context.Context, SnapshotDriver,
	) error {
		return wanted
	}, testScopeVerifier)
	registry := newTestRegistry(t, "forms", testPort(func(
		_ transaction.TenantScope,
		received Program,
	) error {
		_ = received.Run(driver)
		return nil
	}))
	if err := registry.Run(program); !errors.Is(err, wanted) {
		t.Fatalf("outcome=%v", err)
	}
}

func TestRegistryRejectsPortThatSkipsProgramRun(t *testing.T) {
	program := newTestProgram(t, "forms", func(
		context.Context, SnapshotDriver,
	) error {
		t.Fatal("skipped Program executed")
		return nil
	}, testScopeVerifier)
	registry := newTestRegistry(t, "forms", testPort(func(
		transaction.TenantScope, Program,
	) error {
		return nil
	}))
	if err := registry.Run(program); !errors.Is(err, errExecutionContract) {
		t.Fatalf("skipped Run error=%v", err)
	}
}
