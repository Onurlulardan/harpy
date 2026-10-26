package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (store *ExportReconciliationStore) PersistReconciliationSealed(
	ctx context.Context,
	claim app.ExportReconciliationClaim,
	observation app.ExportReconciliationObservation,
	result app.ExportArtifactReconciliation,
	receipt app.ExportReconciliationReceipt,
) error {
	if observation.Validate(claim) != nil || result.State != app.ExportArtifactSealed ||
		result.Validate(observation.ExportEffect) != nil ||
		receipt.Validate(observation, result) != nil {
		return app.ErrExportReconciliationInvalid
	}
	return store.inTenant(ctx, claim, func(scoped context.Context,
		bound *exportReconciliationBound,
	) error {
		schema, err := bound.schema(scoped)
		if err != nil {
			return err
		}
		effect, artifact := observation.ExportEffect, result.Artifact
		arguments := reconciliationObservationArguments(
			schema, claim, observation)
		arguments = append(arguments, artifact.ObjectKey, artifact.ObjectVersion,
			artifact.KeyID, artifact.ProviderRequestID, artifact.EffectID,
			artifact.EffectReceiptID, artifact.HeaderSHA256[:],
			artifact.CiphertextSHA256[:], artifact.RequestSHA256[:],
			artifact.ResultSHA256[:], artifact.CiphertextBytes,
			append([]byte(nil), artifact.WrappedDEK...), artifact.WrapNonce[:],
			artifact.AttemptGeneration, artifact.SealedAt,
			receipt.ProviderRequestID, receipt.ResultSHA256[:], receipt.ObservedAt)
		var replayed bool
		err = bound.database.QueryRow(scoped, reconciliationSealedSQL,
			arguments...).Scan(&replayed)
		if err != nil {
			return mapExportReconciliationError(err)
		}
		if artifact.ObjectKey != effect.ObjectKey {
			return app.ErrExportReconciliationInvalid
		}
		return nil
	})
}
