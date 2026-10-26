package iam

import (
	"context"

	lifecycledeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycledelivery"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *LifecycleDeliveryStateRuntime) Claim(ctx context.Context,
	command lifecycledeliveryapp.ClaimCommand,
) (lifecycledeliveryapp.ClaimDecision, string, error) {
	var result lifecycledeliveryapp.ClaimDecision
	var tenantSlug string
	err := runtime.inTenant(ctx, command.TenantID,
		func(scoped context.Context, scope platformtx.JobTenantScope,
			store lifecycledeliveryapp.Store) error {
			var claimErr error
			result, claimErr = store.Claim(scoped, command)
			tenantSlug = scope.TenantSlug()
			return claimErr
		})
	return result, tenantSlug, err
}
