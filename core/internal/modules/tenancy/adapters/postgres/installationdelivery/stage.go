package installationdelivery

import (
	"context"
	_ "embed"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

//go:embed sql/platform/stage.sql
var stageSQL string

func (storage *store) Stage(
	ctx context.Context, command installationdelivery.StageCommand,
) (installationdelivery.Delivery, error) {
	if !storage.usable(ctx) {
		return installationdelivery.Delivery{}, installationdelivery.ErrUnavailable
	}
	return scanDelivery(storage.queryRow(ctx, stageSQL, command.OperationID,
		command.AttemptID, command.CredentialID, command.CredentialVersion,
		command.InstallationManifestSHA256, command.DeliveryManifestSHA256))
}
