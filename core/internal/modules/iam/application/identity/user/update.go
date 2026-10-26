package user

import (
	"context"
)

func (commands *Commands) UpdateUser(
	ctx context.Context,
	scope TenantScope,
	command UpdateUserCommand,
) (MutationResult, error) {
	if command.Validate() != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	profile, err := commands.profiles.PreparePatch(
		ctx, scope, ProfileUser, command.ID, command.ProfilePatch,
	)
	if err != nil {
		return MutationResult{}, err
	}
	if profile.Validate(ProfileUser) != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	command = command.WithProfile(profile)
	return commands.runMutation(
		ctx, scope, command.Evidence, command.ID, command.ExpectedRevision,
		func(ctx context.Context, scope TenantScope) (MutationResult, error) {
			return commands.store.UpdateUser(ctx, scope, command)
		},
	)
}
