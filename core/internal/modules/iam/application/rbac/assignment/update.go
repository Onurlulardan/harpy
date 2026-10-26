package assignment

import (
	"context"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (commands *Commands) Update(
	ctx context.Context,
	scope platformtx.TenantScope,
	command rbacdomain.AssignmentMutation,
) (rbacdomain.MutationResult, error) {
	if commands == nil || command.ValidateUpdate() != nil ||
		!mutationScope(ctx, scope, command.Evidence) {
		return rbacdomain.MutationResult{}, rbacdomain.ErrInvalidMutation
	}
	result, err := commands.store.Update(ctx, scope, command)
	if err != nil {
		return rbacdomain.MutationResult{}, err
	}
	if result.Validate(command.ID) != nil {
		return rbacdomain.MutationResult{}, rbacdomain.ErrStorage
	}
	return commands.finish(ctx, scope, command.Evidence, result,
		"iam.rbac.assignment.update", "iam.rbac.assignment.updated")
}
