package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (store *exportWorkerBound) BeginArtifactEffect(
	ctx context.Context,
	start app.ExportArtifactEffectStart,
) (bool, error) {
	if !store.usable(ctx, start.Attempt) || start.Validate() != nil {
		return false, app.ErrExportArtifactTampered
	}
	schema, err := store.schema(ctx)
	if err != nil {
		return false, err
	}
	var replayed bool
	err = store.database.QueryRow(ctx, outboundEffectStartSQL,
		outboundEffectArguments(schema, start)...).Scan(&replayed)
	if err != nil {
		return false, mapOutboundError(err, app.ErrExportArtifactTampered)
	}
	return replayed, nil
}
