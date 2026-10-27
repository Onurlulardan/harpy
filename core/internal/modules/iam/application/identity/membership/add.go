package membership

import (
	"context"
)

func (commands *Commands) AddMembership(
	ctx context.Context, scope TenantScope, command AddMembershipCommand,
) (MutationResult, error) {
	if command.Validate() != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	return commands.runMutation(ctx, scope, command.Evidence, command.ID, 0,
		func(ctx context.Context, scope TenantScope) (MutationResult, error) {
			return commands.store.AddMembership(ctx, scope, command)
		})
}
