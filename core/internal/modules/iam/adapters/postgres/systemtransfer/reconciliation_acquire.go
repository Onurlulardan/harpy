package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (store *ExportReconciliationStore) AcquireReconciliation(
	ctx context.Context,
	claim app.ExportReconciliationClaim,
) (app.ExportReconciliationObservation, bool, error) {
	var observation app.ExportReconciliationObservation
	var claimed bool
	err := store.inTenant(ctx, claim, func(scoped context.Context,
		bound *exportReconciliationBound,
	) error {
		schema, err := bound.schema(scoped)
		if err != nil {
			return err
		}
		var raw reconciliationRow
		err = bound.database.QueryRow(scoped, reconciliationAcquireSQL,
			reconciliationClaimArguments(schema, claim)...,
		).Scan(raw.targets()...)
		if err != nil {
			return mapExportReconciliationError(err)
		}
		observation, claimed, err = raw.decode(claim)
		return err
	})
	return observation, claimed, err
}
