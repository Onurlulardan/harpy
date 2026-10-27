package recoveryprotection

import (
	"context"
	"testing"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type backendFixture struct{ called bool }

func (backend *backendFixture) RecoveryProtectionLookup(
	context.Context, string,
) (platformopsapi.RecoveryProtection, error) {
	backend.called = true
	return platformopsapi.RecoveryProtection{EnvironmentID: "environment"}, nil
}

func TestGetForwardsRecoveryProtectionIdentity(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewRecoveryProtectionOperations(backend).GetProtection(
		t.Context(), "environment")
	if err != nil || !backend.called || result.EnvironmentID != "environment" {
		t.Fatalf("GetProtection() = %#v, %v", result, err)
	}
}
