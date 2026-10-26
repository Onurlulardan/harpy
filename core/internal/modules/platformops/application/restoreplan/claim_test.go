package restoreplan

import (
	"context"
	"testing"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type backendFixture struct{ called bool }

func (backend *backendFixture) PlanArtifactClaim(
	context.Context, platformopsapi.RecoveryArtifactClaim,
) (platformopsapi.RecoveryArtifact, error) {
	backend.called = true
	return platformopsapi.RecoveryArtifact{Bytes: []byte("plan")}, nil
}

func TestClaimForwardsRestorePlanIdentity(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewPlanOperations(backend).ClaimRestorePlan(
		t.Context(), platformopsapi.RecoveryArtifactClaim{})
	if err != nil || !backend.called || string(result.Bytes) != "plan" {
		t.Fatalf("ClaimRestorePlan() = %#v, %v", result, err)
	}
}
