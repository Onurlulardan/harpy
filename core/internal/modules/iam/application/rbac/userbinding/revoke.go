package userbinding

import (
	"context"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (commands *Commands) Revoke(ctx context.Context, scope transaction.TenantScope,
	command rbacdomain.BindingCommand,
) (rbacdomain.MutationResult, error) {
	if commands == nil || command.ValidateRevoke() != nil || !mutationScope(ctx, scope, command.Evidence) {
		return rbacdomain.MutationResult{}, rbacdomain.ErrInvalidMutation
	}
	result, err := commands.store.Revoke(ctx, scope, command)
	if err != nil || result.Validate(command.ID) != nil {
		return rbacdomain.MutationResult{}, rbacdomain.ErrStorage
	}
	return commands.finish(ctx, scope, command.Evidence, result,
		"iam.rbac.user_binding.revoke", "iam.rbac.user_binding.revoked")
}
