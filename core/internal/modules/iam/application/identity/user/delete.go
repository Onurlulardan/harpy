package user

import (
	"context"
)

func (commands *Commands) DeleteUser(
	ctx context.Context,
	scope TenantScope,
	command UserStateCommand,
) (MutationResult, error) {
	if command.Validate() != nil || commands == nil ||
		!mutationScopeUsable(ctx, scope, command.Evidence) {
		return MutationResult{}, ErrInvalidCommand
	}
	result, err := commands.store.DeleteUser(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	valid := result.Validate() == nil && result.ID == command.ID
	if result.Replayed {
		valid = valid && result.Revision > command.ExpectedRevision
	} else {
		valid = valid && result.Revision == command.ExpectedRevision+1
	}
	if !valid {
		return MutationResult{}, ErrStorage
	}
	return result, nil
}
