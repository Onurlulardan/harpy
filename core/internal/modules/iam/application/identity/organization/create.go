package organization

import (
	"context"
)

func (commands *Commands) CreateOrganization(
	ctx context.Context,
	scope TenantScope,
	command CreateOrganizationCommand,
) (MutationResult, error) {
	if command.Validate() != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	profile, err := commands.profiles.PrepareNew(
		ctx, scope, ProfileOrganization, command.ProfileValues,
	)
	if err != nil {
		return MutationResult{}, err
	}
	if profile.Validate(ProfileOrganization) != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	command = command.WithProfile(profile)
	return commands.runMutation(ctx, scope, command.Evidence, command.ID, 0,
		func(ctx context.Context, scope TenantScope) (MutationResult, error) {
			return commands.store.CreateOrganization(ctx, scope, command)
		})
}
