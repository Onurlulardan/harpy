package installationdelivery

import (
	"context"
	_ "embed"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

//go:embed sql/platform/complete.sql
var completeSQL string

func (storage *store) complete(
	ctx context.Context, command installationdelivery.StateCommand,
) (installationdelivery.Transition, error) {
	if !storage.usable(ctx) {
		return installationdelivery.Transition{}, installationdelivery.ErrUnavailable
	}
	return scanTransition(storage.queryRow(ctx, completeSQL, command.OperationID,
		command.AttemptID, command.ExpectedRevision))
}
