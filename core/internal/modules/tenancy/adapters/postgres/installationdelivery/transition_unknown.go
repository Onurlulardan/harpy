package installationdelivery

import (
	"context"
	_ "embed"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

//go:embed sql/platform/unknown.sql
var unknownSQL string

func (storage *store) recordUnknownOutcome(
	ctx context.Context, command installationdelivery.UnknownCommand,
) (installationdelivery.Transition, error) {
	if !storage.usable(ctx) {
		return installationdelivery.Transition{}, installationdelivery.ErrUnavailable
	}
	return scanTransition(storage.queryRow(ctx, unknownSQL, command.OperationID,
		command.AttemptID, command.ExpectedRevision, command.AdapterID,
		command.ClaimTokenSHA256, string(command.Cause)))
}
