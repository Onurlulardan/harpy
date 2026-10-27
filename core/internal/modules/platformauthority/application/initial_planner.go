package application

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type InitialPlanner struct{ coordinator *Coordinator }

func NewInitialPlanner(
	store ports.Store,
	signer ports.ProposalSigner,
	clock ports.Clock,
) (*InitialPlanner, error) {
	if store == nil || signer == nil || clock == nil {
		return nil, fmt.Errorf("platform authority planner composition is invalid")
	}
	return &InitialPlanner{coordinator: &Coordinator{
		store: store, signer: signer, clock: clock, registryKind: ports.RegistrySaaS,
	}}, nil
}

func (planner *InitialPlanner) PrepareInitial(
	ctx context.Context,
	request api.InitialAuthorityRequest,
) (api.PreparedAuthority, error) {
	if planner == nil || planner.coordinator == nil {
		return api.PreparedAuthority{}, api.ErrDependency
	}
	return planner.coordinator.PrepareInitial(ctx, request)
}

var _ api.InitialAuthorityPlanner = (*InitialPlanner)(nil)
