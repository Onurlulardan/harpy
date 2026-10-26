package membership

import (
	"context"
)

func (commands *Commands) DeleteMembership(
	ctx context.Context, scope TenantScope, command DeleteMembershipCommand,
) (MutationResult, error) {
	if command.Validate() != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	return commands.runMutation(
		ctx, scope, command.Evidence, command.ID, command.ExpectedRevision,
		func(ctx context.Context, scope TenantScope) (MutationResult, error) {
			return commands.store.DeleteMembership(ctx, scope, command)
		},
	)
}
