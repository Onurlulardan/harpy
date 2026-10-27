package role

import (
	"context"
)

func (commands *Commands) RetireRole(
	ctx context.Context,
	scope TenantScope,
	command RetireRoleCommand,
) (MutationResult, error) {
	if command.Validate() != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	return commands.runMutation(
		ctx, scope, command.Evidence, command.ID, command.ExpectedRevision,
		func(ctx context.Context, scope TenantScope) (MutationResult, error) {
			return commands.store.RetireRole(ctx, scope, command)
		},
	)
}
