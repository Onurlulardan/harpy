package installationdelivery

import (
	"context"
	_ "embed"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

//go:embed sql/platform/claim.sql
var claimSQL string

func (storage *store) claim(
	ctx context.Context, command installationdelivery.ClaimCommand,
) (installationdelivery.Delivery, error) {
	if !storage.usable(ctx) {
		return installationdelivery.Delivery{}, installationdelivery.ErrUnavailable
	}
	return scanDelivery(storage.queryRow(ctx, claimSQL, command.OperationID,
		command.AttemptID, command.ExpectedRevision, command.AdapterID,
		command.ClaimTokenSHA256, command.LeaseSeconds))
}
