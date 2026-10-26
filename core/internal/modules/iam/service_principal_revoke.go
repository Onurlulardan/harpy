package iam

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *ServicePrincipalRuntime) Revoke(
	ctx context.Context,
	scope platformtx.TenantScope,
	command ServicePrincipalRevokeCommand,
) (ServicePrincipalMutationResult, error) {
	if !runtime.Valid() {
		return ServicePrincipalMutationResult{}, ErrServicePrincipalRuntime
	}
	return runtime.commands.Revoke(ctx, scope, command)
}
