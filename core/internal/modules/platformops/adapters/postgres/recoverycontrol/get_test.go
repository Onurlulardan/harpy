package recoverycontrol

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type backendFixture struct{ called bool }

func (backend *backendFixture) RecoveryControlLookup(
	context.Context, string,
) (ports.RecoveryControlOperationRecord, error) {
	backend.called = true
	return ports.RecoveryControlOperationRecord{OperationID: "operation"}, nil
}

func TestGetForwardsRecoveryControlIdentity(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewRecoveryControlOperations(backend).GetRecoveryControlOperation(
		t.Context(), "operation")
	if err != nil || !backend.called || result.OperationID != "operation" {
		t.Fatalf("GetRecoveryControlOperation() = %#v, %v", result, err)
	}
}
