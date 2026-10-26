package recoverykit

import (
	"context"
	"testing"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type backendFixture struct{ called bool }

func (backend *backendFixture) RecoveryKitClaim(
	context.Context, string, string, string,
) (platformopsapi.RecoveryKitClaim, error) {
	backend.called = true
	return platformopsapi.RecoveryKitClaim{ArtifactBytes: []byte("kit")}, nil
}

func TestClaimForwardsRecoveryKitIdentity(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewRecoveryKitOperations(backend).ClaimKit(
		t.Context(), "operation", "actor", "session")
	if err != nil || !backend.called || string(result.ArtifactBytes) != "kit" {
		t.Fatalf("ClaimKit() = %#v, %v", result, err)
	}
}
