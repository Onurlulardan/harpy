package user

import (
	"context"
)

func (commands *Commands) CreateUser(
	ctx context.Context,
	scope TenantScope,
	command CreateUserCommand,
) (MutationResult, error) {
	if command.Validate() != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	profile, err := commands.profiles.PrepareNew(
		ctx, scope, ProfileUser, command.ProfileValues,
	)
	if err != nil {
		return MutationResult{}, err
	}
	if profile.Validate(ProfileUser) != nil {
		return MutationResult{}, ErrInvalidCommand
	}
	command = command.WithProfile(profile)
	return commands.runMutation(ctx, scope, command.Evidence, command.ID, 0,
		func(ctx context.Context, scope TenantScope) (MutationResult, error) {
			return commands.store.CreateUser(ctx, scope, command)
		})
}
