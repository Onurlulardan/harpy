package application

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (service *AirGapCoordinator) PrepareInitialExport(
	ctx context.Context,
	request api.InitialAuthorityRequest,
) (api.AirGapRequestArtifact, error) {
	state, err := artifact.CanonicalizeState(artifactState(request.State))
	if err != nil || request.Policy != api.ExpiryAirGap {
		return api.AirGapRequestArtifact{}, api.ErrConflict
	}
	key := ports.HeadKey{
		RegistryBindingID: request.State.RegistryBindingID,
		ScopeKind:         artifact.ScopeKind(request.State.ScopeKind), ScopeID: request.State.ScopeID,
	}
	purpose := "initial_environment_authority"
	if request.State.ScopeKind == api.ScopeTenant {
		purpose = "initial_tenant_authority"
	}
	command := operationCommand{
		operationID: request.OperationID, purpose: purpose, state: state,
		generation: 1, issuedAt: request.IssuedAt, policy: artifact.ExpiryAirGap,
	}
	if _, err := service.base.store.LoadOperation(ctx, request.OperationID); err == nil {
		operation, prepareErr := service.base.prepareOperation(ctx, command)
		if prepareErr != nil {
			return api.AirGapRequestArtifact{}, prepareErr
		}
		return airGapArtifact(operation), nil
	} else if !errors.Is(err, ports.ErrNotFound) {
		return api.AirGapRequestArtifact{}, api.ErrDependency
	}
	if _, err := service.base.store.LoadHead(ctx, key); err == nil {
		return api.AirGapRequestArtifact{}, api.ErrStaleHead
	} else if !errors.Is(err, ports.ErrNotFound) {
		return api.AirGapRequestArtifact{}, api.ErrDependency
	}
	operation, err := service.base.prepareOperation(ctx, command)
	if err != nil {
		return api.AirGapRequestArtifact{}, err
	}
	return airGapArtifact(operation), nil
}

func airGapArtifact(operation ports.OperationRecord) api.AirGapRequestArtifact {
	return api.AirGapRequestArtifact{
		OperationID:    operation.OperationID,
		ProposalSHA256: operation.Proposal.SHA256,
		Bytes:          append([]byte(nil), operation.Proposal.Bytes...),
		ExpiresAt:      operation.Proposal.Value.ExpiresAt,
	}
}
