package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *RecoveryService) RecoverySetListing(
	ctx context.Context,
	environmentID string,
) ([]api.RecoverySet, error) {
	if service == nil || ctx == nil || ctx.Err() != nil ||
		!validRecoveryUUID(environmentID) {
		return nil, api.ErrInvalid
	}
	records, err := service.sets.ListRecoverySets(ctx, environmentID)
	if err != nil {
		return nil, mapRecoveryStoreError(err)
	}
	result := make([]api.RecoverySet, len(records))
	for index, record := range records {
		result[index] = recoverySetAPI(record)
	}
	return result, nil
}
