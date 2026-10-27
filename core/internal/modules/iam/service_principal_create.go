package iam

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *ServicePrincipalRuntime) Create(
	ctx context.Context,
	scope platformtx.TenantScope,
	command ServicePrincipalCreateCommand,
) (ServicePrincipalMutationResult, error) {
	if !runtime.Valid() {
		return ServicePrincipalMutationResult{}, ErrServicePrincipalRuntime
	}
	return runtime.commands.Create(ctx, scope, command)
}
