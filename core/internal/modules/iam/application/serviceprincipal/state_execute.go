package serviceprincipal

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type stateOperation func(
	context.Context, transaction.TenantScope, StateCommand,
) (MutationResult, error)

func (commands *Commands) changeState(
	ctx context.Context,
	scope transaction.TenantScope,
	command StateCommand,
	operation stateOperation,
	commandKey string,
	resultCode string,
) (MutationResult, error) {
	if commands == nil || operation == nil || command.Validate() != nil ||
		!mutationScopeUsable(ctx, scope, command.Evidence) {
		return MutationResult{}, ErrInvalidPrincipal
	}
	result, err := operation(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	if !validMutationResult(result, command.ID) ||
		result.Revision != command.ExpectedRevision+1 {
		return MutationResult{}, ErrEvidenceMissing
	}
	return commands.finish(
		ctx, scope, command.Evidence, result, commandKey, resultCode,
	)
}
