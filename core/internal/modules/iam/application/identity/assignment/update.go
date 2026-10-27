package assignment

import (
	"context"
)

func (commands *Commands) UpdateAssignment(
	ctx context.Context, scope TenantScope, command UpdateAssignmentCommand,
) (MutationResult, error) {
	if command.Validate() != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	return commands.runMutation(
		ctx, scope, command.Evidence, command.ID, command.ExpectedRevision,
		func(ctx context.Context, scope TenantScope) (MutationResult, error) {
			return commands.store.UpdateAssignment(ctx, scope, command)
		},
	)
}
