package resource

import (
	"context"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (commands *Commands) Update(
	ctx context.Context, scope transaction.TenantScope, command rbacdomain.UpdateCatalog,
) (rbacdomain.MutationResult, error) {
	if commands == nil || command.Validate() != nil || !mutationScope(ctx, scope, command.Evidence) {
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
		"iam.rbac.resource.update", "iam.rbac.resource.updated")
}
