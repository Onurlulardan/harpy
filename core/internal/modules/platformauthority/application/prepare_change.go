package application

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (service *Coordinator) PrepareChange(
	ctx context.Context,
	request api.ProtectedChangeRequest,
) (api.PreparedAuthority, error) {
	purpose := string(request.Purpose)
	state, err := artifact.CanonicalizeState(artifactState(request.State))
	if err != nil || request.OperationID == "" || !protectedChangePurpose(purpose) {
		return api.PreparedAuthority{}, api.ErrConflict
	}
	if existing, loadErr := service.store.LoadOperation(ctx, request.OperationID); loadErr == nil {
		command := operationCommand{operationID: request.OperationID, purpose: purpose,
			state: state, generation: existing.Proposal.Value.Generation,
			previousHead: existing.Proposal.Value.PreviousHeadSHA256,
			issuedAt:     request.IssuedAt, policy: artifactExpiry(request.Policy)}
		operation, prepareErr := service.prepareOperation(ctx, command)
		if prepareErr != nil {
			return api.PreparedAuthority{}, prepareErr
		}
		return preparedAuthority(operation), nil
	} else if !errors.Is(loadErr, ports.ErrNotFound) {
		return api.PreparedAuthority{}, api.ErrDependency
	}
	key := ports.HeadKey{RegistryBindingID: request.State.RegistryBindingID,
		ScopeKind: artifact.ScopeKind(request.State.ScopeKind), ScopeID: request.State.ScopeID}
	head, err := service.store.LoadHead(ctx, key)
	if err != nil {
		return api.PreparedAuthority{}, mapHeadLoadError(err)
	}
	if head.StateSHA256 == state.SHA256 {
		return api.PreparedAuthority{}, api.ErrConflict
	}
	previous := head.AckSHA256
	operation, err := service.prepareOperation(ctx, operationCommand{
		operationID: request.OperationID, purpose: purpose, state: state,
		generation: head.Generation + 1, previousHead: &previous,
		issuedAt: request.IssuedAt, policy: artifactExpiry(request.Policy),
	})
	if err != nil {
		return api.PreparedAuthority{}, err
	}
	return preparedAuthority(operation), nil
}

func protectedChangePurpose(value string) bool {
	return value == string(api.ChangeOwnerTransfer) ||
		value == string(api.ChangeProtectedAdmin) ||
		value == string(api.ChangeRecoveryAnchor)
}

func mapHeadLoadError(err error) error {
	if errors.Is(err, ports.ErrNotFound) || errors.Is(err, ports.ErrStaleHead) {
		return api.ErrStaleHead
	}
	return api.ErrDependency
}
