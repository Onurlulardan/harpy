package application

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (service *Coordinator) CompensateUncommitted(
	ctx context.Context,
	request api.CompensationRequest,
) (api.AuthorityReceipt, error) {
	state, err := artifact.CanonicalizeState(artifactState(request.State))
	if err != nil || request.OperationID == "" || request.CompensatesAckSHA256 == "" {
		return api.AuthorityReceipt{}, api.ErrConflict
	}
	key := ports.HeadKey{
		RegistryBindingID: request.State.RegistryBindingID,
		ScopeKind:         artifact.ScopeKind(request.State.ScopeKind), ScopeID: request.State.ScopeID,
	}
	if existing, err := service.store.LoadOperation(ctx, request.OperationID); err == nil {
		compensates := request.CompensatesAckSHA256
		return service.runOperation(ctx, operationCommand{
			operationID: request.OperationID, purpose: "compensate_uncommitted_ack",
			state: state, generation: existing.Proposal.Value.Generation,
			previousHead:   existing.Proposal.Value.PreviousHeadSHA256,
			compensatesAck: &compensates, issuedAt: request.IssuedAt,
			policy: artifactExpiry(request.Policy),
		})
	} else if !errors.Is(err, ports.ErrNotFound) {
		return api.AuthorityReceipt{}, api.ErrDependency
	}
	head, err := service.store.LoadHead(ctx, key)
	if errors.Is(err, ports.ErrNotFound) ||
		(err == nil && head.AckSHA256 != request.CompensatesAckSHA256) {
		return api.AuthorityReceipt{}, api.ErrStaleHead
	}
	if err != nil {
		return api.AuthorityReceipt{}, api.ErrDependency
	}
	previous := head.AckSHA256
	compensates := request.CompensatesAckSHA256
	return service.runOperation(ctx, operationCommand{
		operationID: request.OperationID, purpose: "compensate_uncommitted_ack",
		state: state, generation: head.Generation + 1, previousHead: &previous,
		compensatesAck: &compensates, issuedAt: request.IssuedAt,
		policy: artifactExpiry(request.Policy),
	})
}
