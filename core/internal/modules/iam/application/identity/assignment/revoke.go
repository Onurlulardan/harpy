package assignment

import (
	"context"
)

func (commands *Commands) RevokeAssignment(
	ctx context.Context, scope TenantScope, command RevokeAssignmentCommand,
) (MutationResult, error) {
	if command.Validate() != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	return commands.runMutation(
		ctx, scope, command.Evidence, command.ID, command.ExpectedRevision,
		func(ctx context.Context, scope TenantScope) (MutationResult, error) {
			return commands.store.RevokeAssignment(ctx, scope, command)
		},
	)
}
