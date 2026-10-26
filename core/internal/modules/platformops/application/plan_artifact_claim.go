package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *RecoveryService) PlanArtifactClaim(
	ctx context.Context,
	claim api.RecoveryArtifactClaim,
) (api.RecoveryArtifact, error) {
	if service == nil || ctx == nil || ctx.Err() != nil || !validArtifactClaim(claim) {
		return api.RecoveryArtifact{}, api.ErrInvalid
	}
	record, err := service.restores.ClaimRestorePlan(ctx, ports.RecoveryArtifactClaim{
		ArtifactID: claim.ArtifactID, TenantID: claim.TenantID,
		ActorID: claim.ActorID, SessionID: claim.SessionID, ClaimedAt: service.clock.Now(),
	})
	if err != nil {
		return api.RecoveryArtifact{}, mapRecoveryStoreError(err)
	}
	return recoveryArtifactAPI(record)
}
