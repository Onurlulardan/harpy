package application

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type Coordinator struct {
	store        ports.Store
	registry     ports.Registry
	signer       ports.ProposalSigner
	clock        ports.Clock
	trust        artifact.VerifiedTrustBundle
	registryKind artifact.RegistryKind
}

func NewCoordinator(
	store ports.Store,
	registry ports.Registry,
	signer ports.ProposalSigner,
	clock ports.Clock,
	trust artifact.VerifiedTrustBundle,
	registryKind artifact.RegistryKind,
) (*Coordinator, error) {
	if store == nil || registry == nil || signer == nil || clock == nil ||
		trust.Kind != registryKind {
		return nil, fmt.Errorf("platform authority composition is invalid")
	}
	return &Coordinator{
		store: store, registry: registry, signer: signer, clock: clock,
		trust: trust, registryKind: registryKind,
	}, nil
}
