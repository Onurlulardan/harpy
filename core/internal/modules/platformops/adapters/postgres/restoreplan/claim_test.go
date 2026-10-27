package restoreplan

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type backendFixture struct{ called bool }

func (backend *backendFixture) PlanArtifactClaim(
	context.Context, ports.RecoveryArtifactClaim,
) (ports.RecoveryArtifactRecord, error) {
	backend.called = true
	return ports.RecoveryArtifactRecord{Bytes: []byte("plan")}, nil
}

func TestClaimForwardsRestorePlanIdentity(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewPlanOperations(backend).ClaimRestorePlan(
		t.Context(), ports.RecoveryArtifactClaim{})
	if err != nil || !backend.called || string(result.Bytes) != "plan" {
		t.Fatalf("ClaimRestorePlan() = %#v, %v", result, err)
	}
}
