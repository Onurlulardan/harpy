package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (store *ExportReconciliationStore) PersistReconciliationMissing(
	ctx context.Context,
	claim app.ExportReconciliationClaim,
	observation app.ExportReconciliationObservation,
	result app.ExportArtifactReconciliation,
	receipt app.ExportReconciliationReceipt,
	code string,
) error {
	if observation.Validate(claim) != nil || result.State != app.ExportArtifactMissing ||
		result.Validate(observation.ExportEffect) != nil ||
		receipt.Validate(observation, result) != nil ||
		code != app.ExportReconciliationMissingCode {
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
		arguments = append(arguments, receipt.ProviderRequestID,
			receipt.ResultSHA256[:], receipt.ObservedAt, code)
		var replayed bool
		err = bound.database.QueryRow(scoped, reconciliationMissingSQL,
			arguments...).Scan(&replayed)
		return mapExportReconciliationError(err)
	})
}
