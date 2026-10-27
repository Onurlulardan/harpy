package user

import (
	"context"
)

func (commands *Commands) DisableUser(
	ctx context.Context,
	scope TenantScope,
	command UserStateCommand,
) (MutationResult, error) {
	if command.Validate() != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	return commands.runMutation(
		ctx, scope, command.Evidence, command.ID, command.ExpectedRevision,
		func(ctx context.Context, scope TenantScope) (MutationResult, error) {
			return commands.store.DisableUser(ctx, scope, command)
		},
	)
}
