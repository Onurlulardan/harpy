package installationdelivery

import (
	"context"
	_ "embed"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

//go:embed sql/platform/acknowledge.sql
var acknowledgeSQL string

func (storage *store) Acknowledge(
	ctx context.Context, command installationdelivery.AcknowledgeCommand,
) (installationdelivery.Delivery, error) {
	if !storage.usable(ctx) {
		return installationdelivery.Delivery{}, installationdelivery.ErrUnavailable
	}
	return scanDelivery(storage.queryRow(ctx, acknowledgeSQL, command.OperationID,
		command.AttemptID, command.ExpectedRevision, command.AdapterID,
		command.ClaimTokenSHA256))
}
