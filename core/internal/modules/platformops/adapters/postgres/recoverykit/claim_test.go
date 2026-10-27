package recoverykit

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type backendFixture struct{ called bool }

func (backend *backendFixture) RecoveryKitClaim(
	context.Context, string, string, string, time.Time,
) (ports.RecoveryKitClaimRecord, error) {
	backend.called = true
	return ports.RecoveryKitClaimRecord{ArtifactBytes: []byte("kit")}, nil
}

func TestClaimForwardsRecoveryKitIdentity(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewRecoveryKitOperations(backend).ClaimRecoveryKit(
		t.Context(), "operation", "actor", "session", time.Now())
	if err != nil || !backend.called || string(result.ArtifactBytes) != "kit" {
		t.Fatalf("ClaimRecoveryKit() = %#v, %v", result, err)
	}
}
