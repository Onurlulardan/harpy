package application

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type InitialAcknowledger struct{ coordinator *Coordinator }

func NewInitialAcknowledger(
	store ports.Store,
	registry ports.Registry,
	clock ports.Clock,
	trust artifact.VerifiedTrustBundle,
) (*InitialAcknowledger, error) {
	if store == nil || registry == nil || clock == nil || trust.Kind != ports.RegistrySaaS {
		return nil, fmt.Errorf("platform authority acknowledger composition is invalid")
	}
	return &InitialAcknowledger{coordinator: &Coordinator{
		store: store, registry: registry, clock: clock, trust: trust,
		registryKind: ports.RegistrySaaS,
	}}, nil
}

func (acknowledger *InitialAcknowledger) AcknowledgePrepared(
	ctx context.Context,
	operationID string,
) (api.AuthorityReceipt, error) {
	if acknowledger == nil || acknowledger.coordinator == nil {
		return api.AuthorityReceipt{}, api.ErrDependency
	}
	return acknowledger.coordinator.AcknowledgePrepared(ctx, operationID)
}

var _ api.InitialAuthorityAcknowledger = (*InitialAcknowledger)(nil)
