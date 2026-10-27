package billing

import (
	"context"

	usagepostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/usage"
	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type documentsUsageEffects struct{}

func (documentsUsageEffects) PublishThreshold(
	context.Context,
	usageapp.ThresholdEventIntent,
) error {
	return usageapp.ErrConfiguration
}

func (service *DocumentsUsage) bindActorCommands(
	scope platformtx.TenantScope,
) (*usageapp.Commands, error) {
	if service == nil {
		return nil, usageapp.ErrConfiguration
	}
	store, err := BindUsageStore(scope, service.catalogs)
	if err != nil {
		return nil, err
	}
	return usageapp.NewCommands(store, service.nextID, service.clock)
}

func (service *DocumentsUsage) bindJobStore(
	scope platformtx.JobTenantScope,
) (usageapp.Store, error) {
	if service == nil {
		return nil, usageapp.ErrConfiguration
	}
	return usagepostgres.BindJob(scope, documentsUsageEffects{}, service.jobActorID)
}

func (service *DocumentsUsage) bindJobCommands(
	scope platformtx.JobTenantScope,
) (*usageapp.Commands, error) {
	store, err := service.bindJobStore(scope)
	if err != nil {
		return nil, err
	}
	return usageapp.NewCommands(store, service.nextID, service.clock)
}
