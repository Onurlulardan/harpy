package platformauthority

import (
	"fmt"

	postgresadapter "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/providerregistry"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func NewSaaSAcknowledger(
	dependencies SaaSAcknowledgerDependencies,
) (api.InitialAuthorityAcknowledger, error) {
	if dependencies.Clock == nil {
		return nil, fmt.Errorf("SaaS platform authority acknowledger clock is missing")
	}
	trust, err := ports.VerifyTrustBundle(dependencies.AckTrustBundle,
		ports.RegistrySaaS, dependencies.Clock.Now(), dependencies.ReleaseTrustRoot)
	if err != nil {
		return nil, fmt.Errorf("verify SaaS authority trust: %w", err)
	}
	registry, err := providerregistry.New(dependencies.RegistryURL, dependencies.HTTPClient)
	if err != nil {
		return nil, err
	}
	store, err := postgresadapter.NewStore(dependencies.Pool, dependencies.Clock)
	if err != nil {
		return nil, err
	}
	return application.NewInitialAcknowledger(store, registry, dependencies.Clock, trust)
}
