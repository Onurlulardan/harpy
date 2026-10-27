package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *RecoveryProtectionService) RecoveryProtectionLookup(
	ctx context.Context,
	environmentID string,
) (api.RecoveryProtection, error) {
	if service == nil || !validRecoveryUUID(environmentID) || ctx == nil || ctx.Err() != nil {
		return api.RecoveryProtection{}, api.ErrInvalid
	}
	record, err := service.store.GetRecoveryProtection(ctx, environmentID)
	if err != nil {
		return api.RecoveryProtection{}, mapRecoveryStoreError(err)
	}
	return api.RecoveryProtection{
		EnvironmentID: record.EnvironmentID, State: record.State,
		Generation: record.Generation, KitAcknowledged: record.KitAcknowledged,
		UpdatedAt: record.UpdatedAt,
	}, nil
}
