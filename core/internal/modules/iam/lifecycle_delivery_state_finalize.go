package iam

import (
	"context"

	lifecycledeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycledelivery"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *LifecycleDeliveryStateRuntime) Finalize(ctx context.Context,
	command lifecycledeliveryapp.FinalizeCommand,
) (lifecycledeliveryapp.StoredResult, error) {
	var result lifecycledeliveryapp.StoredResult
	err := runtime.inTenant(ctx, command.TenantID,
		func(scoped context.Context, _ platformtx.JobTenantScope,
			store lifecycledeliveryapp.Store) error {
			var finalizeErr error
			result, finalizeErr = store.Finalize(scoped, command)
			return finalizeErr
		})
	return result, err
}
