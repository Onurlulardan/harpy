package installationdelivery

import (
	"context"
	_ "embed"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

//go:embed sql/platform/mark_secret_deleted.sql
var markSecretDeletedSQL string

func (storage *store) deleteSecret(
	ctx context.Context, command installationdelivery.StateCommand,
) (installationdelivery.Transition, error) {
	if !storage.usable(ctx) {
		return installationdelivery.Transition{}, installationdelivery.ErrUnavailable
	}
	return scanTransition(storage.queryRow(ctx, markSecretDeletedSQL,
		command.OperationID, command.AttemptID, command.ExpectedRevision))
}
