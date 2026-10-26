package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *RecoveryService) RecoveryManifestClaim(
	ctx context.Context,
	claim api.RecoveryArtifactClaim,
) (api.RecoveryArtifact, error) {
	if service == nil || ctx == nil || ctx.Err() != nil || !validArtifactClaim(claim) {
		return api.RecoveryArtifact{}, api.ErrInvalid
	}
	record, err := service.sets.ClaimRecoveryManifest(ctx, ports.RecoveryArtifactClaim{
		ArtifactID: claim.ArtifactID, TenantID: claim.TenantID,
		ActorID: claim.ActorID, SessionID: claim.SessionID, ClaimedAt: service.clock.Now(),
	})
	if err != nil {
		return api.RecoveryArtifact{}, mapRecoveryStoreError(err)
	}
	return recoveryArtifactAPI(record)
}

func validArtifactClaim(claim api.RecoveryArtifactClaim) bool {
	return validRecoveryUUID(claim.ArtifactID) && validRecoveryUUID(claim.TenantID) &&
		validRecoveryUUID(claim.ActorID) && validRecoveryUUID(claim.SessionID)
}

func recoveryArtifactAPI(record ports.RecoveryArtifactRecord) (api.RecoveryArtifact, error) {
	if !validRecoveryDigest(record.SHA256) || len(record.Bytes) == 0 {
		return api.RecoveryArtifact{}, api.ErrConflict
	}
	return api.RecoveryArtifact{SHA256: record.SHA256,
		Bytes: append([]byte(nil), record.Bytes...)}, nil
}
