package operator

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	installationdeliverypostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/installationdelivery"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func buildInstallationDelivery(
	pool *platformpostgres.Pool,
) (*tenancy.InstallationDeliveryRuntime, error) {
	store, err := installationdeliverypostgres.NewCommandStore(pool)
	if err != nil {
		return nil, err
	}
	recovery, err := installationdeliverypostgres.NewRecoveryUnitOfWork(pool)
	if err != nil {
		return nil, err
	}
	return tenancy.NewInstallationDeliveryRuntime(store, recovery)
}
