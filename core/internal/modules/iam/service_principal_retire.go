package iam

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *ServicePrincipalRuntime) Retire(
	ctx context.Context,
	scope platformtx.TenantScope,
	command ServicePrincipalStateCommand,
) (ServicePrincipalMutationResult, error) {
	if !runtime.Valid() {
		return ServicePrincipalMutationResult{}, ErrServicePrincipalRuntime
	}
	return runtime.commands.Retire(ctx, scope, command)
}
