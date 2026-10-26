package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *RecoveryService) GetRecoverySet(
	ctx context.Context,
	recoverySetID string,
) (api.RecoverySet, error) {
	if service == nil || ctx == nil || ctx.Err() != nil || recoverySetID == "" {
		return api.RecoverySet{}, api.ErrInvalid
	}
	record, err := service.sets.GetRecoverySet(ctx, recoverySetID)
	if err != nil {
		return api.RecoverySet{}, mapRecoveryStoreError(err)
	}
	return recoverySetAPI(record), nil
}

func recoverySetAPI(record ports.RecoverySetRecord) api.RecoverySet {
	return api.RecoverySet{RecoverySetID: record.RecoverySetID,
		OperationID: record.OperationID, EnvironmentID: record.EnvironmentID,
		Sequence: record.Sequence, State: record.State,
		ManifestSHA256: record.ManifestSHA256, LegalHold: record.LegalHold,
		DatabaseSystemID: record.DatabaseSystemID, ProtectedAt: record.ProtectedAt}
}
