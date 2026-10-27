package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (store *exportWorkerBound) PersistOutboundSuccess(
	ctx context.Context,
	completion app.ExportCompletion,
) (bool, error) {
	attempt, progress, artifact := completion.Attempt,
		completion.Progress, completion.Artifact
	if !store.usable(ctx, attempt) || progress.Validate() != nil ||
		artifact.Validate() != nil || artifact.AttemptGeneration != attempt.AttemptGeneration {
		return false, app.ErrExportArtifactTampered
	}
	schema, err := store.schema(ctx)
	if err != nil {
		return false, err
	}
	arguments := append(outboundAttemptArguments(schema, attempt),
		progress.Rows, progress.PlaintextBytes, progress.CheckedAt,
		artifact.ObjectKey, artifact.ObjectVersion, artifact.KeyID,
		artifact.ProviderRequestID, artifact.EffectID, artifact.EffectReceiptID,
		artifact.HeaderSHA256[:], artifact.CiphertextSHA256[:],
		artifact.RequestSHA256[:], artifact.ResultSHA256[:],
		artifact.CiphertextBytes, append([]byte(nil), artifact.WrappedDEK...),
		artifact.WrapNonce[:], artifact.SealedAt)
	var replayed bool
	err = store.database.QueryRow(ctx, outboundCompleteSQL,
		arguments...).Scan(&replayed)
	if err != nil {
		return false, mapOutboundError(err, app.ErrExportArtifactTampered)
	}
	return replayed, nil
}
