package recoverymanifest

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type backendFixture struct{ called bool }

func (backend *backendFixture) RecoveryManifestClaim(
	context.Context, ports.RecoveryArtifactClaim,
) (ports.RecoveryArtifactRecord, error) {
	backend.called = true
	return ports.RecoveryArtifactRecord{Bytes: []byte("manifest")}, nil
}

func TestClaimForwardsRecoveryManifestIdentity(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewRecoveryManifestOperations(backend).ClaimRecoveryManifest(
		t.Context(), ports.RecoveryArtifactClaim{})
	if err != nil || !backend.called || string(result.Bytes) != "manifest" {
		t.Fatalf("ClaimRecoveryManifest() = %#v, %v", result, err)
	}
}
