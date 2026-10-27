package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (store *ExportReconciliationStore) PersistReconciliationParked(
	ctx context.Context,
	claim app.ExportReconciliationClaim,
	observation app.ExportReconciliationObservation,
	record app.ExportReconciliationUnknownRecord,
) error {
	if observation.Validate(claim) != nil || record.Validate(claim, observation) != nil ||
		record.Code != app.ExportReconciliationParkedCode {
		return app.ErrExportReconciliationInvalid
	}
	return store.inTenant(ctx, claim, func(scoped context.Context,
		bound *exportReconciliationBound,
	) error {
		schema, err := bound.schema(scoped)
		if err != nil {
			return err
		}
		arguments := reconciliationObservationArguments(
			schema, claim, observation)
		arguments = append(arguments, record.FailureSHA256[:], record.ObservedAt,
			record.Code)
		var replayed bool
		err = bound.database.QueryRow(scoped, reconciliationParkedSQL,
			arguments...).Scan(&replayed)
		return mapExportReconciliationError(err)
	})
}
