package permission

import (
	"context"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (commands *Commands) Create(ctx context.Context, scope transaction.TenantScope,
	input rbacdomain.PermissionInput,
) (rbacdomain.MutationResult, error) {
	if commands == nil || input.Validate() != nil || !mutationScope(ctx, scope, input.Evidence) {
		return rbacdomain.MutationResult{}, rbacdomain.ErrInvalidMutation
	}
	result, err := commands.store.Create(ctx, scope, input)
	if err != nil {
		return rbacdomain.MutationResult{}, err
	}
	if result.Validate(input.ID) != nil {
		return rbacdomain.MutationResult{}, rbacdomain.ErrStorage
	}
	return commands.finish(ctx, scope, input.Evidence, result,
		"iam.rbac.permission.create", "iam.rbac.permission.created")
}
