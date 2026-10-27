package recoveryexecution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *Service) Complete(
	ctx context.Context,
	record ports.RecoveryCompletionRecord,
) error {
	if service == nil || ctx == nil || ctx.Err() != nil ||
		!validIdentity(ports.RecoveryExecutionIdentity{
			RecoverySetID: record.RecoverySetID,
			AttemptID:     record.AttemptID,
			LeaseToken:    record.LeaseToken,
		}) || !validDigest(record.ManifestSHA256) ||
		len(record.ManifestBytes) < 1 || len(record.ManifestBytes) > 2*1024*1024 ||
		!validUUID(record.DatabaseSystemID) {
		return ErrInvalid
	}
	changed, err := service.store.CompleteRecoverySet(ctx, record)
	if err != nil {
		return ErrUnavailable
	}
	if !changed {
		return ErrConflict
	}
	return nil
}
