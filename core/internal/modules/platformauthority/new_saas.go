package platformauthority

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/gcpkms"
	postgresadapter "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/providerregistry"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func NewSaaS(
	ctx context.Context,
	dependencies SaaSDependencies,
) (Runtime, error) {
	if dependencies.Clock == nil {
		return Runtime{}, fmt.Errorf("SaaS platform authority clock is missing")
	}
	trust, err := ports.VerifyTrustBundle(
		dependencies.AckTrustBundle, ports.RegistrySaaS,
		dependencies.Clock.Now(), dependencies.ReleaseTrustRoot,
	)
	if err != nil {
		return Runtime{}, fmt.Errorf("verify SaaS authority trust: %w", err)
	}
	signer, err := gcpkms.NewSigner(
		ctx, dependencies.KMSKeyID, dependencies.KMSPublicKeySHA256,
		dependencies.KMSClient,
	)
	if err != nil {
		return Runtime{}, err
	}
	registry, err := providerregistry.New(dependencies.RegistryURL, dependencies.HTTPClient)
	if err != nil {
		return Runtime{}, err
	}
	store, err := postgresadapter.NewStore(dependencies.Pool, dependencies.Clock)
	if err != nil {
		return Runtime{}, err
	}
	binding, err := application.NewBindingService(
		store, registry, signer, dependencies.Clock, ports.RegistrySaaS,
	)
	if err != nil {
		return Runtime{}, err
	}
	initial, err := application.NewCoordinator(
		store, registry, signer, dependencies.Clock, trust, ports.RegistrySaaS,
	)
	if err != nil {
		return Runtime{}, err
	}
	return Runtime{Binding: binding, Initial: initial, Protected: initial, Signer: initial}, nil
}
