package passwordenrollment

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type Commands struct{ store Store }

func NewCommands(store Store) (*Commands, error) {
	if nilInterface(store) {
		return nil, ErrStorage
	}
	return &Commands{store: store}, nil
}

func (commands *Commands) Create(
	ctx context.Context,
	scope TenantScope,
	command Command,
) (MutationResult, error) {
	if commands == nil || command.Validate() != nil || ctx == nil ||
		ctx.Err() != nil || platformtx.NetworkAllowed(ctx) ||
		!platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, platformtx.Command) != nil ||
		scope.ActorID() != uuid.UUID(command.Evidence.ActorID) {
		return MutationResult{}, ErrInvalidCommand
	}
	result, err := commands.store.Create(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	if result.Validate() != nil || result.ID != command.UserID || result.Revision != 1 {
		return MutationResult{}, ErrStorage
	}
	return result, nil
}
