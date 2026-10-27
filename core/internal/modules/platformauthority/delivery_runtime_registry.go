package platformauthority

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"os"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/devjournal"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/providerregistry"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func newDeliveryRegistry(
	_ context.Context,
	dependencies DeliveryDependencies,
	trust ports.VerifiedTrustBundle,
) (ports.Registry, error) {
	switch dependencies.Mode {
	case DeliveryDevelopment:
		return newDevelopmentDeliveryRegistry(dependencies, trust)
	case DeliverySaaS:
		client, err := providerregistry.NewAuthenticatedHTTPClient(
			dependencies.WorkloadIdentityPath,
			dependencies.HasWorkloadIdentity,
		)
		if err != nil {
			return nil, err
		}
		return providerregistry.New(dependencies.RegistryURL, client)
	default:
		return nil, fmt.Errorf("platform authority delivery mode is invalid")
	}
}

func newDevelopmentDeliveryRegistry(
	dependencies DeliveryDependencies,
	trust ports.VerifiedTrustBundle,
) (ports.Registry, error) {
	seed, err := os.ReadFile(dependencies.LocalAckSeedPath)
	if err != nil || len(seed) != ed25519.SeedSize {
		clear(seed)
		return nil, fmt.Errorf("development authority seed is invalid")
	}
	defer clear(seed)
	private := ed25519.NewKeyFromSeed(seed)
	defer clear(private)
	kid, public, found := trust.ActiveKey()
	if !found || !private.Public().(ed25519.PublicKey).Equal(public) {
		return nil, fmt.Errorf("development authority signer does not match trust")
	}
	resolver, err := devjournal.NewRecoveryBundleResolver(
		dependencies.RecoveryTrustPath,
	)
	if err != nil {
		return nil, err
	}
	return devjournal.New(
		dependencies.LocalJournalPath, kid, private, resolver,
		dependencies.Clock,
	)
}
