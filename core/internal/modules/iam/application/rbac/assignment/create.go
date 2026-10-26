package assignment

import (
	"context"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (commands *Commands) Create(
	ctx context.Context,
	scope platformtx.TenantScope,
	command rbacdomain.AssignmentMutation,
) (rbacdomain.MutationResult, error) {
	if commands == nil || command.ValidateCreate() != nil ||
		!mutationScope(ctx, scope, command.Evidence) {
		return rbacdomain.MutationResult{}, rbacdomain.ErrInvalidMutation
	}
	result, err := commands.store.Create(ctx, scope, command)
	if err != nil {
		return rbacdomain.MutationResult{}, err
	}
	if result.Validate(command.ID) != nil {
		return rbacdomain.MutationResult{}, rbacdomain.ErrStorage
	}
	return commands.finish(ctx, scope, command.Evidence, result,
		"iam.rbac.assignment.create", "iam.rbac.assignment.created")
}
