package systemtransfer

import (
	"context"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

type ExportReconciliationRecoveryResult struct {
	requeued int
	parked   int
}

func (result ExportReconciliationRecoveryResult) Requeued() int {
	return result.requeued
}

func (result ExportReconciliationRecoveryResult) Parked() int {
	return result.parked
}

func (result ExportReconciliationRecoveryResult) Total() int {
	return result.requeued + result.parked
}

func (store *ExportReconciliationStore) RecoverExpiredExportReconciliations(
	ctx context.Context,
	tenantID uuid.UUID,
	limit int,
) (ExportReconciliationRecoveryResult, error) {
	if store == nil || store.runInTenant == nil || ctx == nil || ctx.Err() != nil ||
		tenantID == uuid.Nil || limit < 1 || limit > jobsv1.MaximumClaimBatch {
		return ExportReconciliationRecoveryResult{},
			app.ErrExportReconciliationInvalid
	}
	var result ExportReconciliationRecoveryResult
	err := store.runInTenant(ctx, tenantID, func(scoped context.Context,
		bound *exportReconciliationBound,
	) error {
		schema, err := bound.schema(scoped)
		if err != nil {
			return err
		}
		for range limit {
			var recovered bool
			var resolution *string
			err = bound.database.QueryRow(scoped,
				recoverExpiredExportReconciliationSQL, schema, tenantID,
			).Scan(&recovered, &resolution)
			if err != nil {
				return mapExportReconciliationError(err)
			}
			if !recovered {
				if resolution != nil {
					return app.ErrExportStateConflict
				}
				break
			}
			if resolution == nil {
				return app.ErrExportStateConflict
			}
			switch *resolution {
			case "requeued":
				result.requeued++
			case "parked":
				result.parked++
			default:
				return app.ErrExportStateConflict
			}
		}
		return nil
	})
	return result, err
}
