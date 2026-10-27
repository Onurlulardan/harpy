package restoreplan

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type Backend interface {
	PlanArtifactClaim(context.Context, platformopsapi.RecoveryArtifactClaim) (platformopsapi.RecoveryArtifact, error)
}

type RestorePlanOperations struct{ backend Backend }

func NewPlanOperations(backend Backend) *RestorePlanOperations {
	return &RestorePlanOperations{backend: backend}
}

func (operations *RestorePlanOperations) ClaimRestorePlan(
	ctx context.Context, claim platformopsapi.RecoveryArtifactClaim,
) (platformopsapi.RecoveryArtifact, error) {
	return operations.backend.PlanArtifactClaim(ctx, claim)
}
