package assignment

import (
	"context"
)

func (commands *Commands) GrantAssignment(
	ctx context.Context, scope TenantScope, command GrantAssignmentCommand,
) (MutationResult, error) {
	if command.Validate() != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	return commands.runMutation(ctx, scope, command.Evidence, command.ID, 0,
		func(ctx context.Context, scope TenantScope) (MutationResult, error) {
			return commands.store.GrantAssignment(ctx, scope, command)
		})
}
