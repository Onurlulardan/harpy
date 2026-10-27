package application

import (
	"context"
	"errors"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (service *Coordinator) PrepareInitial(
	ctx context.Context,
	request api.InitialAuthorityRequest,
) (api.PreparedAuthority, error) {
	state, err := artifact.CanonicalizeState(artifactState(request.State))
	if err != nil || request.OperationID == "" {
		return api.PreparedAuthority{}, api.ErrConflict
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
		generation: 1, issuedAt: request.IssuedAt, policy: artifactExpiry(request.Policy),
	}
	if _, err := service.store.LoadOperation(ctx, request.OperationID); err == nil {
		operation, prepareErr := service.prepareOperation(ctx, command)
		if prepareErr != nil {
			return api.PreparedAuthority{}, prepareErr
		}
		return preparedAuthority(operation), nil
	} else if !errors.Is(err, ports.ErrNotFound) {
		return api.PreparedAuthority{}, api.ErrDependency
	}
	if _, err := service.store.LoadHead(ctx, key); err == nil {
		return api.PreparedAuthority{}, api.ErrStaleHead
	} else if !errors.Is(err, ports.ErrNotFound) {
		return api.PreparedAuthority{}, api.ErrDependency
	}
	operation, err := service.prepareOperation(ctx, command)
	if err != nil {
		return api.PreparedAuthority{}, err
	}
	return preparedAuthority(operation), nil
}

func preparedAuthority(operation ports.OperationRecord) api.PreparedAuthority {
	expiresAt, _ := time.Parse(time.RFC3339, operation.Proposal.Value.ExpiresAt)
	return api.PreparedAuthority{
		OperationID:       operation.OperationID,
		RegistryBindingID: operation.Proposal.Value.RegistryBindingID,
		ScopeKind:         string(operation.Proposal.Value.ScopeKind),
		ScopeID:           operation.Proposal.Value.ScopeID,
		Generation:        operation.Proposal.Value.Generation,
		StateSHA256:       operation.State.SHA256,
		ProposalSHA256:    operation.Proposal.SHA256,
		ProposalBytes:     append([]byte(nil), operation.Proposal.Bytes...),
		ExpiresAt:         expiresAt,
	}
}
