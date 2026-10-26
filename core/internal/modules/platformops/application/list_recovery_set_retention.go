package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *RecoveryService) ListRecoverySetRetention(
	ctx context.Context,
	environmentID string,
) ([]api.RecoveryRetentionCandidate, error) {
	if service == nil || ctx == nil || ctx.Err() != nil || environmentID == "" {
		return nil, api.ErrInvalid
	}
	records, err := service.sets.ListRetentionCandidates(ctx, environmentID)
	if err != nil {
		return nil, mapRecoveryStoreError(err)
	}
	result := make([]api.RecoveryRetentionCandidate, len(records))
	for index, record := range records {
		result[index] = api.RecoveryRetentionCandidate{
			RecoverySetID: record.RecoverySetID, ManifestSHA256: record.ManifestSHA256,
			ProtectedAt: record.ProtectedAt, LegalHold: record.LegalHold,
		}
	}
	return result, nil
}
