package recoveryanchor

import (
	"context"

	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) Register(
	ctx context.Context,
	scope platformtx.TenantScope,
	command anchordomain.RegisterCommand,
) (anchordomain.ProposedMutation, error) {
	if service == nil || !scopeValid(scope, command.Subject) {
		return anchordomain.ProposedMutation{}, anchordomain.ErrInvalid
	}
	staged, err := service.store.StageRegistration(ctx, scope, command)
	if err != nil {
		return anchordomain.ProposedMutation{}, err
	}
	prepared, err := service.authority.Prepare(
		ctx, command.OperationID, staged.Authority, command.OccurredAt)
	if err != nil {
		return anchordomain.ProposedMutation{}, anchordomain.ErrDependency
	}
	if err = service.store.PersistProposal(
		ctx, scope, command.Subject, prepared, command.OccurredAt); err != nil {
		return anchordomain.ProposedMutation{}, err
	}
	return anchordomain.ProposedMutation{OperationID: command.OperationID,
		AnchorID: staged.AnchorID, Kind: "anchor_register", State: "accepted"}, nil
}
