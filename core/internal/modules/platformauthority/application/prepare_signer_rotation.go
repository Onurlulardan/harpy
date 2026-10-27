package application

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (service *Coordinator) PrepareSignerRotation(
	ctx context.Context,
	request api.SignerRotationRequest,
) (api.PreparedAuthority, error) {
	state, err := artifact.CanonicalizeState(artifactState(request.State))
	if err != nil || request.OperationID == "" || request.State.ScopeKind != api.ScopeEnvironment {
		return api.PreparedAuthority{}, api.ErrConflict
	}
	if existing, loadErr := service.store.LoadOperation(ctx, request.OperationID); loadErr == nil {
		return service.replaySignerRotation(ctx, request, state, existing)
	} else if !errors.Is(loadErr, ports.ErrNotFound) {
		return api.PreparedAuthority{}, api.ErrDependency
	}
	binding, err := service.store.LoadBinding(ctx, request.State.RegistryBindingID)
	if err != nil || !bindingMatchesRotation(binding, request, service.signer.KID()) {
		return api.PreparedAuthority{}, mapStoreError(errOrConflict(err))
	}
	evidence, valid := buildRotationEvidence(request, binding)
	if !valid {
		return api.PreparedAuthority{}, api.ErrConflict
	}
	key := ports.HeadKey{RegistryBindingID: request.State.RegistryBindingID,
		ScopeKind: artifact.ScopeEnvironment, ScopeID: request.State.ScopeID}
	head, err := service.store.LoadHead(ctx, key)
	if err != nil {
		return api.PreparedAuthority{}, mapHeadLoadError(err)
	}
	previous := head.AckSHA256
	operation, err := service.prepareOperation(ctx, operationCommand{
		operationID: request.OperationID, purpose: "proposal_signer_rotation",
		state: state, generation: head.Generation + 1, previousHead: &previous,
		rotation: evidence, issuedAt: request.IssuedAt,
		policy: artifactExpiry(request.Policy),
	})
	if err != nil {
		return api.PreparedAuthority{}, err
	}
	return preparedAuthority(operation), nil
}

func bindingMatchesRotation(
	binding ports.BindingRecord,
	request api.SignerRotationRequest,
	activeSignerKID string,
) bool {
	return binding.BindingID == request.State.RegistryBindingID &&
		binding.EnvironmentID == request.State.EnvironmentID &&
		string(binding.DeploymentMode) == string(request.State.DeploymentMode) &&
		binding.SignerStatus == "active" && binding.SignerKID == activeSignerKID
}

func errOrConflict(err error) error {
	if err == nil {
		return ports.ErrConflict
	}
	return err
}
