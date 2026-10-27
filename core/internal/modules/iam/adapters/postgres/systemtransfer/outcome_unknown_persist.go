package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (store *exportWorkerBound) PersistOutcomeUnknown(
	ctx context.Context,
	record app.ExportOutcomeUnknownRecord,
) error {
	if !store.usable(ctx, record.Effect.Attempt) || record.Validate() != nil {
		return app.ErrExportWorkerInvalid
	}
	schema, err := store.schema(ctx)
	if err != nil {
		return err
	}
	arguments := outboundEffectArguments(schema, record.Effect)
	var objectVersion, providerRequestID any
	var resultSHA, sealedAt any
	if record.ArtifactKnown {
		objectVersion, providerRequestID = record.Artifact.ObjectVersion,
			record.Artifact.ProviderRequestID
		resultSHA, sealedAt = record.Artifact.ResultSHA256[:], record.Artifact.SealedAt
	}
	arguments = append(arguments, record.ArtifactKnown, objectVersion,
		providerRequestID, resultSHA, sealedAt, record.At, record.Code)
	var replayed bool
	err = store.database.QueryRow(ctx, outboundUnknownSQL,
		arguments...).Scan(&replayed)
	if err != nil {
		return mapOutboundError(err, app.ErrExportArtifactTampered)
	}
	return nil
}
