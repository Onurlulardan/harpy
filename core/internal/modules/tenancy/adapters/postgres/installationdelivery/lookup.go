package installationdelivery

import (
	"context"
	_ "embed"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

//go:embed sql/platform/read.sql
var readSQL string

func (storage *store) read(
	ctx context.Context, query installationdelivery.ReadQuery,
) (installationdelivery.Delivery, error) {
	if !storage.usable(ctx) {
		return installationdelivery.Delivery{}, installationdelivery.ErrUnavailable
	}
	return scanDelivery(storage.queryRow(ctx, readSQL, query.OperationID))
}
