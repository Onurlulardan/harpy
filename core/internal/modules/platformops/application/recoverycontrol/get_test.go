package recoverycontrol

import (
	"context"
	"testing"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type backendFixture struct{ called bool }

func (backend *backendFixture) RecoveryControlLookup(
	context.Context, string,
) (platformopsapi.RecoveryControlOperation, error) {
	backend.called = true
	return platformopsapi.RecoveryControlOperation{OperationID: "operation"}, nil
}

func TestGetForwardsRecoveryControlIdentity(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewRecoveryControlOperations(backend).GetControlOperation(
		t.Context(), "operation")
	if err != nil || !backend.called || result.OperationID != "operation" {
		t.Fatalf("GetControlOperation() = %#v, %v", result, err)
	}
}
