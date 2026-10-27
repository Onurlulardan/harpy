package execution

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRegistrySurfacesOwnerCleanupFailureAfterSuccessfulProgram(t *testing.T) {
	cleanupErr := errors.New("owner cleanup failed")
	program := newTestProgram(t, "forms", func(
		context.Context, SnapshotDriver,
	) error {
		return nil
	}, testScopeVerifier)
	registry := newTestRegistry(t, "forms", testPort(func(
		_ transaction.TenantScope,
		received Program,
	) error {
		if err := received.Run(&testDriver{owner: "forms"}); err != nil {
			return err
		}
		return cleanupErr
	}))
	if err := registry.Run(program); !errors.Is(err, cleanupErr) {
		t.Fatalf("cleanup error disappeared: %v", err)
	}
}

func TestRegistryPreservesProgramAndCleanupFailures(t *testing.T) {
	programErr := errors.New("program failed")
	cleanupErr := errors.New("owner cleanup failed")
	program := newTestProgram(t, "forms", func(
		context.Context, SnapshotDriver,
	) error {
		return programErr
	}, testScopeVerifier)
	registry := newTestRegistry(t, "forms", testPort(func(
		_ transaction.TenantScope,
		received Program,
	) error {
		return errors.Join(received.Run(&testDriver{owner: "forms"}), cleanupErr)
	}))
	err := registry.Run(program)
	if !errors.Is(err, programErr) || !errors.Is(err, cleanupErr) {
		t.Fatalf("joined outcome changed: %v", err)
	}
}
