package execution

import (
	"context"
	"errors"
	"testing"
	"unsafe"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestProgramRejectsSameOwnerDriverFromForeignTenantScope(t *testing.T) {
	foreign := distinctOpaqueScopeForComparison(t)
	closureCalls := 0
	program := newTestProgram(t, "forms", func(
		context.Context, SnapshotDriver,
	) error {
		closureCalls++
		return nil
	}, testScopeVerifier)
	registry := newTestRegistry(t, "forms", testPort(func(
		_ transaction.TenantScope,
		received Program,
	) error {
		return received.Run(&testDriver{owner: "forms", scope: foreign})
	}))
	if err := registry.Run(program); !errors.Is(err, errExecutionContract) {
		t.Fatalf("foreign scope driver error=%v", err)
	}
	if closureCalls != 0 {
		t.Fatal("foreign scope driver reached closure")
	}
}

// TenantScope deliberately has no public constructor. This adversarial test
// changes only its opaque pointer identity and never dereferences the forged
// state; production comparison remains ordinary TenantScope equality.
func distinctOpaqueScopeForComparison(t *testing.T) transaction.TenantScope {
	t.Helper()
	var scope transaction.TenantScope
	token := new([64]byte)
	*(*unsafe.Pointer)(unsafe.Pointer(&scope)) = unsafe.Pointer(token)
	if scope == (transaction.TenantScope{}) {
		t.Fatal("foreign opaque scope identity was not created")
	}
	return scope
}
