package restoreplan

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type Backend interface {
	PlanArtifactClaim(context.Context, ports.RecoveryArtifactClaim) (ports.RecoveryArtifactRecord, error)
}

type RestorePlanOperations struct{ backend Backend }

func NewPlanOperations(backend Backend) *RestorePlanOperations {
	return &RestorePlanOperations{backend: backend}
}

func (operations *RestorePlanOperations) ClaimRestorePlan(
	ctx context.Context, claim ports.RecoveryArtifactClaim,
) (ports.RecoveryArtifactRecord, error) {
	return operations.backend.PlanArtifactClaim(ctx, claim)
}
