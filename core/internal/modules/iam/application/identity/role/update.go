package role

import (
	"context"
)

func (commands *Commands) UpdateRole(
	ctx context.Context,
	scope TenantScope,
	command UpdateRoleCommand,
) (MutationResult, error) {
	if command.Validate() != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	profile, err := commands.profiles.PreparePatch(
		ctx, scope, ProfileRole, command.ID, command.ProfilePatch,
	)
	if err != nil {
		return MutationResult{}, err
	}
	if profile.Validate(ProfileRole) != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	command = command.WithProfile(profile)
	return commands.runMutation(
		ctx, scope, command.Evidence, command.ID, command.ExpectedRevision,
		func(ctx context.Context, scope TenantScope) (MutationResult, error) {
			return commands.store.UpdateRole(ctx, scope, command)
		},
	)
}
