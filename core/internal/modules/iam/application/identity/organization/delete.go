package organization

import (
	"context"
)

func (commands *Commands) DeleteOrganization(
	ctx context.Context,
	scope TenantScope,
	command OrganizationStateCommand,
) (MutationResult, error) {
	if command.Validate() != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	return commands.runMutation(
		ctx, scope, command.Evidence, command.ID, command.ExpectedRevision,
		func(ctx context.Context, scope TenantScope) (MutationResult, error) {
			return commands.store.DeleteOrganization(ctx, scope, command)
		},
	)
}
