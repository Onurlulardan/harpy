package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *RecoveryProtectionService) RecoveryControlLookup(
	ctx context.Context,
	operationID string,
) (api.RecoveryControlOperation, error) {
	if service == nil || ctx == nil || ctx.Err() != nil || !validRecoveryUUID(operationID) {
		return api.RecoveryControlOperation{}, api.ErrInvalid
	}
	record, err := service.store.GetRecoveryControlOperation(ctx, operationID)
	if err != nil {
		return api.RecoveryControlOperation{}, mapRecoveryStoreError(err)
	}
	return recoveryControlView(record), nil
}

func recoveryControlView(
	record ports.RecoveryControlOperationRecord,
) api.RecoveryControlOperation {
	return api.RecoveryControlOperation{
		OperationID: record.OperationID, EnvironmentID: record.EnvironmentID,
		Kind: record.Kind, State: record.State, KitClaimState: record.KitClaimState,
		ArtifactSHA256: record.ArtifactSHA256, FailureCode: record.FailureCode,
		ExpiresAt: record.ExpiresAt, CreatedAt: record.CreatedAt,
	}
}
