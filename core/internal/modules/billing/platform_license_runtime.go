package billing

import (
	"context"
	"fmt"

	platformlicensepostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/platformlicense"
	platformlicenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformlicense"
	"github.com/google/uuid"
)

type PlatformLicenseRuntimeDatabase interface {
	RuntimeOwnerDatabase()
}

type PlatformLicenseApplication interface {
	Read(context.Context, uuid.UUID) (platformlicenseapp.License, error)
	Create(
		context.Context,
		platformlicenseapp.RefreshCommand,
	) (platformlicenseapp.RefreshAccepted, error)
	Get(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) (platformlicenseapp.RefreshOperation, error)
}

func NewPlatformLicenseRuntime(
	database PlatformLicenseRuntimeDatabase,
) (PlatformLicenseApplication, error) {
	store, err := platformlicensepostgres.BindAPI(database)
	if err != nil {
		return nil, fmt.Errorf("bind billing platform license store: %w", err)
	}
	service, err := platformlicenseapp.NewService(store)
	if err != nil {
		return nil, fmt.Errorf("build billing platform license service: %w", err)
	}
	return service, nil
}
