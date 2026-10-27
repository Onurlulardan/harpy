package tenancy

import (
	"context"

	tenancypostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/recoverytarget"
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	recoveryapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/recoverytarget"
)

type recoveryTargetsFacade struct {
	service *recoveryapp.Service
}

func NewRecoveryTargets(
	database any,
	purpose string,
) (tenantapi.RecoveryTargets, error) {
	store, err := tenancypostgres.Bind(database)
	if err != nil {
		return nil, err
	}
	service, err := recoveryapp.New(store, purpose)
	if err != nil {
		return nil, err
	}
	return &recoveryTargetsFacade{service: service}, nil
}

func (facade *recoveryTargetsFacade) Get(
	ctx context.Context,
) (tenantapi.RecoveryTarget, error) {
	return facade.service.Get(ctx)
}
