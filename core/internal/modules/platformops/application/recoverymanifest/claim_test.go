package recoverymanifest

import (
	"context"
	"testing"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type backendFixture struct{ called bool }

func (backend *backendFixture) RecoveryManifestClaim(
	context.Context, platformopsapi.RecoveryArtifactClaim,
) (platformopsapi.RecoveryArtifact, error) {
	backend.called = true
	return platformopsapi.RecoveryArtifact{Bytes: []byte("manifest")}, nil
}

func TestClaimForwardsRecoveryManifestIdentity(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewRecoveryManifestOperations(backend).ClaimRecoveryManifest(
		t.Context(), platformopsapi.RecoveryArtifactClaim{})
	if err != nil || !backend.called || string(result.Bytes) != "manifest" {
		t.Fatalf("ClaimRecoveryManifest() = %#v, %v", result, err)
	}
}
