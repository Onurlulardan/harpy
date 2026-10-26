package organization

import (
	"context"
)

func (commands *Commands) UpdateOrganization(
	ctx context.Context,
	scope TenantScope,
	command UpdateOrganizationCommand,
) (MutationResult, error) {
	if command.Validate() != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	profile, err := commands.profiles.PreparePatch(
		ctx, scope, ProfileOrganization, command.ID, command.ProfilePatch,
	)
	if err != nil {
		return MutationResult{}, err
	}
	if profile.Validate(ProfileOrganization) != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	command = command.WithProfile(profile)
	return commands.runMutation(
		ctx, scope, command.Evidence, command.ID, command.ExpectedRevision,
		func(ctx context.Context, scope TenantScope) (MutationResult, error) {
			return commands.store.UpdateOrganization(ctx, scope, command)
		},
	)
}
