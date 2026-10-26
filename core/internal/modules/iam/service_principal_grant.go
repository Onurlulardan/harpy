package iam

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *ServicePrincipalRuntime) Grant(
	ctx context.Context,
	scope platformtx.TenantScope,
	command ServicePrincipalGrantCommand,
) (ServicePrincipalMutationResult, error) {
	if !runtime.Valid() {
		return ServicePrincipalMutationResult{}, ErrServicePrincipalRuntime
	}
	return runtime.commands.Grant(ctx, scope, command)
}
