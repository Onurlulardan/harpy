package application

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type AirGapCoordinator struct {
	base       *Coordinator
	authorizer ports.AirGapAuthorizer
}

var _ api.AirGapAuthority = (*AirGapCoordinator)(nil)

func NewAirGapCoordinator(
	store ports.Store,
	signer ports.ProposalSigner,
	clock ports.Clock,
	trust artifact.VerifiedTrustBundle,
	authorizer ports.AirGapAuthorizer,
) (*AirGapCoordinator, error) {
	if store == nil || signer == nil || clock == nil || authorizer == nil ||
		trust.Kind != artifact.RegistryDevelopment {
		return nil, fmt.Errorf("air-gap authority composition is invalid")
	}
	return &AirGapCoordinator{
		base: &Coordinator{
			store: store, signer: signer, clock: clock, trust: trust,
			registryKind: artifact.RegistryDevelopment,
		},
		authorizer: authorizer,
	}, nil
}
