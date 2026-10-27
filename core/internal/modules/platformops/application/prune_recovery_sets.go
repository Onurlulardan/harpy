package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
)

func (service *RecoveryService) PruneRecoverySets(
	ctx context.Context,
	environmentID string,
) (api.RetentionResult, error) {
	if service == nil || ctx == nil || ctx.Err() != nil || environmentID == "" {
		return api.RetentionResult{}, api.ErrInvalid
	}
	records, err := service.ListRecoverySetRetention(ctx, environmentID)
	if err != nil {
		return api.RetentionResult{}, mapRecoveryStoreError(err)
	}
	candidates := make([]recoveryset.RetentionCandidate, len(records))
	digests := make(map[string]string, len(records))
	for index, record := range records {
		candidates[index] = recoveryset.RetentionCandidate{ID: record.RecoverySetID,
			ProtectedAt: record.ProtectedAt, LegalHold: record.LegalHold}
		digests[record.RecoverySetID] = record.ManifestSHA256
	}
	decision := recoveryset.ApplyRetention(candidates)
	for _, recoverySetID := range decision.Delete {
		if err = service.MarkRecoverySetDeleting(
			ctx, recoverySetID, digests[recoverySetID],
		); err != nil {
			return api.RetentionResult{}, err
		}
	}
	return api.RetentionResult{Keep: decision.Keep, Delete: decision.Delete}, nil
}
