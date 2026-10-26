package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *RecoveryService) ResolveRestore(
	ctx context.Context,
	operationID string,
) (api.RestoreOperation, error) {
	if service == nil || ctx == nil || ctx.Err() != nil || operationID == "" {
		return api.RestoreOperation{}, api.ErrInvalid
	}
	record, err := service.restores.ResolveRestore(ctx, operationID)
	if err != nil {
		return api.RestoreOperation{}, mapRecoveryStoreError(err)
	}
	return viewFromRecord(record), nil
}

func viewFromRecord(record ports.RestoreRecord) api.RestoreOperation {
	return api.RestoreOperation{OperationID: record.OperationID,
		RecoverySetID: record.RecoverySetID, EnvironmentID: record.EnvironmentID,
		Disposition: api.TargetDisposition(record.Disposition),
		State:       record.State, PlanSHA256: record.PlanSHA256,
		ReceiptSHA256:       record.ReceiptSHA256,
		NewDatabaseSystemID: record.NewDatabaseSystemID,
		CreatedAt:           record.CreatedAt, UpdatedAt: record.UpdatedAt,
		CompletedAt: record.CompletedAt}
}
