package permission

import (
	"context"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (commands *Commands) Retire(ctx context.Context, scope transaction.TenantScope,
	command rbacdomain.StateCommand,
) (rbacdomain.MutationResult, error) {
	if commands == nil || command.Validate() != nil || !mutationScope(ctx, scope, command.Evidence) {
		return rbacdomain.MutationResult{}, rbacdomain.ErrInvalidMutation
	}
	result, err := commands.store.Retire(ctx, scope, command)
	if err != nil {
		return rbacdomain.MutationResult{}, err
	}
	if result.Validate(command.ID) != nil {
		return rbacdomain.MutationResult{}, rbacdomain.ErrStorage
	}
	return commands.finish(ctx, scope, command.Evidence, result,
		"iam.rbac.permission.retire", "iam.rbac.permission.retired")
}
