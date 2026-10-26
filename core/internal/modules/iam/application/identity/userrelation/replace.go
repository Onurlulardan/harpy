package userrelation

import (
	"context"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

func (commands *Commands) Replace(
	ctx context.Context,
	scope TenantScope,
	command ReplaceCommand,
) (MutationResult, error) {
	if commands == nil || command.Validate() != nil ||
		!mutationScopeUsable(ctx, scope, command.Evidence) {
		return MutationResult{}, ErrInvalidCommand
	}
	if command.PatchProfile {
		profile, err := commands.profiles.PreparePatch(
			ctx, scope, identitydomain.ProfileUser,
			command.UserID, command.ProfilePatch,
		)
		if err != nil {
			return MutationResult{}, err
		}
		if profile.Validate(identitydomain.ProfileUser) != nil {
			return MutationResult{}, ErrInvalidCommand
		}
		command = command.WithProfile(profile)
	}
	result, err := commands.store.ReplaceUserRelations(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	if result.Validate() != nil || result.ID != command.UserID ||
		!result.Replayed && result.Revision != command.ExpectedRevision+1 {
		return MutationResult{}, ErrStorage
	}
	return result, nil
}
