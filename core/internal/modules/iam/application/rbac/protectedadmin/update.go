package protectedadmin

import (
	"context"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (commands *Commands) Update(ctx context.Context, scope transaction.TenantScope,
	command rbacdomain.UpdatePolicyCommand) (rbacdomain.MutationResult, error) {
	if commands == nil || command.Validate() != nil || !mutationScope(ctx, scope, command.Evidence) {
		return rbacdomain.MutationResult{}, rbacdomain.ErrInvalidMutation
	}
	result, err := commands.store.Update(ctx, scope, command)
	if err != nil || result.Validate(command.PolicyID) != nil {
		return rbacdomain.MutationResult{}, rbacdomain.ErrStorage
	}
	if result.Replayed {
		return result, nil
	}
	change := rbacdomain.NewChange(result, command.Evidence,
		"iam.rbac.protected_admin.update", "iam.rbac.protected_admin.updated")
	if change.Validate() != nil || commands.sink.WriteChange(ctx, scope, change) != nil {
		return rbacdomain.MutationResult{}, rbacdomain.ErrSink
	}
	return result, nil
}
