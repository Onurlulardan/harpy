package recoveryprotection

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type backendFixture struct{ called bool }

func (backend *backendFixture) RecoveryProtectionLookup(
	context.Context, string,
) (ports.RecoveryProtectionRecord, error) {
	backend.called = true
	return ports.RecoveryProtectionRecord{EnvironmentID: "environment"}, nil
}

func TestGetForwardsRecoveryProtectionIdentity(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewRecoveryProtectionOperations(backend).GetRecoveryProtection(
		t.Context(), "environment")
	if err != nil || !backend.called || result.EnvironmentID != "environment" {
		t.Fatalf("GetRecoveryProtection() = %#v, %v", result, err)
	}
}
