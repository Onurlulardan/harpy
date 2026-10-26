package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *RecoveryProtectionService) RecoveryKitClaim(
	ctx context.Context,
	operationID, actorUserID, sessionID string,
) (api.RecoveryKitClaim, error) {
	if service == nil || ctx == nil || ctx.Err() != nil ||
		!validRecoveryUUID(operationID) || !validRecoveryUUID(actorUserID) ||
		!validRecoveryUUID(sessionID) {
		return api.RecoveryKitClaim{}, api.ErrInvalid
	}
	record, err := service.store.ClaimRecoveryKit(
		ctx, operationID, actorUserID, sessionID, service.clock.Now())
	if err != nil {
		return api.RecoveryKitClaim{}, mapRecoveryStoreError(err)
	}
	if len(record.ArtifactBytes) == 0 || !validRecoveryDigest(record.ArtifactSHA256) ||
		record.ExpiresAt.IsZero() {
		return api.RecoveryKitClaim{}, api.ErrConflict
	}
	return api.RecoveryKitClaim{
		ArtifactBytes:  append([]byte(nil), record.ArtifactBytes...),
		ArtifactSHA256: record.ArtifactSHA256, ExpiresAt: record.ExpiresAt,
	}, nil
}
