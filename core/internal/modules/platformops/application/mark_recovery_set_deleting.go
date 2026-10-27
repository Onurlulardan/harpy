package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *RecoveryService) MarkRecoverySetDeleting(
	ctx context.Context,
	recoverySetID string,
	manifestSHA256 string,
) error {
	if service == nil || ctx == nil || ctx.Err() != nil ||
		recoverySetID == "" || !validRecoveryDigest(manifestSHA256) {
		return api.ErrInvalid
	}
	if err := service.sets.MarkRecoverySetDeleting(
		ctx, recoverySetID, manifestSHA256, service.clock.Now(),
	); err != nil {
		return mapRecoveryStoreError(err)
	}
	return nil
}
