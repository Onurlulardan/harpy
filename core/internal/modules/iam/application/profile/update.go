package profile

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (commands *Commands) UpdateSelf(
	ctx context.Context,
	scope transaction.TenantScope,
	command SelfUpdateCommand,
) (MutationResult, error) {
	if commands == nil || command.Validate() != nil ||
		!selfUpdateScopeUsable(ctx, scope, command) {
		return MutationResult{}, ErrInvalidCommand
	}
	result, err := commands.port.UpdateSelf(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	if !validResult(result, command.ActorID) {
		return MutationResult{}, ErrEvidenceMissing
	}
	return result, nil
}

func selfUpdateScopeUsable(
	ctx context.Context,
	scope transaction.TenantScope,
	command SelfUpdateCommand,
) bool {
	return ctx != nil && ctx.Err() == nil && !transaction.NetworkAllowed(ctx) &&
		transaction.ContextOwnsScope(ctx, scope) &&
		transaction.RequireActive(scope, transaction.Command) == nil &&
		scope.ActorID() == command.ActorID
}
