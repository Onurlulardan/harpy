package iam

import (
	"context"

	lifecycledeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycledelivery"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *LifecycleDeliveryStateRuntime) MarkUnknown(ctx context.Context,
	reference lifecycledeliveryapp.EffectReference) error {
	return runtime.inTenant(ctx, reference.TenantID,
		func(scoped context.Context, _ platformtx.JobTenantScope,
			store lifecycledeliveryapp.Store) error {
			return store.MarkUnknown(scoped, reference)
		})
}
