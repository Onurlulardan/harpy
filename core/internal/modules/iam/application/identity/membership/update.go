package membership

import (
	"context"
)

func (commands *Commands) UpdateMembership(
	ctx context.Context, scope TenantScope, command UpdateMembershipCommand,
) (MutationResult, error) {
	if command.Validate() != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	return commands.runMutation(
		ctx, scope, command.Evidence, command.ID, command.ExpectedRevision,
		func(ctx context.Context, scope TenantScope) (MutationResult, error) {
			return commands.store.UpdateMembership(ctx, scope, command)
		},
	)
}
