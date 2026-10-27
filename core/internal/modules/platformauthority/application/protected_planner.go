package application

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type ProtectedPlanner struct{ coordinator *Coordinator }

func NewProtectedPlanner(
	store ports.Store,
	signer ports.ProposalSigner,
	clock ports.Clock,
) (*ProtectedPlanner, error) {
	if store == nil || signer == nil || clock == nil {
		return nil, fmt.Errorf("protected authority planner composition is invalid")
	}
	return &ProtectedPlanner{coordinator: &Coordinator{
		store: store, signer: signer, clock: clock,
	}}, nil
}

func (planner *ProtectedPlanner) PrepareChange(
	ctx context.Context,
	request api.ProtectedChangeRequest,
) (api.PreparedAuthority, error) {
	if planner == nil || planner.coordinator == nil {
		return api.PreparedAuthority{}, api.ErrDependency
	}
	return planner.coordinator.PrepareChange(ctx, request)
}

var _ api.ProtectedAuthorityPlanner = (*ProtectedPlanner)(nil)
