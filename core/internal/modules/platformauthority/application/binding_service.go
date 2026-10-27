package application

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type BindingService struct {
	store        ports.Store
	registry     ports.Registry
	signer       ports.ProposalSigner
	clock        ports.Clock
	registryKind artifact.RegistryKind
}

func NewBindingService(
	store ports.Store,
	registry ports.Registry,
	signer ports.ProposalSigner,
	clock ports.Clock,
	registryKind artifact.RegistryKind,
) (*BindingService, error) {
	if store == nil || registry == nil || signer == nil || clock == nil ||
		(registryKind != artifact.RegistryDevelopment && registryKind != artifact.RegistrySaaS) {
		return nil, fmt.Errorf("platform authority binding composition is invalid")
	}
	return &BindingService{
		store: store, registry: registry, signer: signer, clock: clock,
		registryKind: registryKind,
	}, nil
}
