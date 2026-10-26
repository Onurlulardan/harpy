package organizationbinding

import (
	"context"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (commands *Commands) Grant(ctx context.Context, scope transaction.TenantScope,
	command rbacdomain.BindingCommand) (rbacdomain.MutationResult, error) {
	if commands == nil || command.ValidateGrant() != nil || !mutationScope(ctx, scope, command.Evidence) {
		return rbacdomain.MutationResult{}, rbacdomain.ErrInvalidMutation
	}
	result, err := commands.store.Grant(ctx, scope, command)
	if err != nil || result.Validate(command.ID) != nil {
		return rbacdomain.MutationResult{}, rbacdomain.ErrStorage
	}
	return commands.finish(ctx, scope, command.Evidence, result,
		"iam.rbac.organization_binding.grant", "iam.rbac.organization_binding.granted")
}
