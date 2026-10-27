package role

import (
	"context"
)

func (commands *Commands) CreateRole(
	ctx context.Context,
	scope TenantScope,
	command CreateRoleCommand,
) (MutationResult, error) {
	if command.Validate() != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	profile, err := commands.profiles.PrepareNew(
		ctx, scope, ProfileRole, command.ProfileValues,
	)
	if err != nil {
		return MutationResult{}, err
	}
	if profile.Validate(ProfileRole) != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	command = command.WithProfile(profile)
	return commands.runMutation(ctx, scope, command.Evidence, command.ID, 0,
		func(ctx context.Context, scope TenantScope) (MutationResult, error) {
			return commands.store.CreateRole(ctx, scope, command)
		})
}
