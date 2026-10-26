package iam

import (
	"context"
	"errors"

	lifecycledeliverypostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/lifecycledelivery"
	lifecycledeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycledelivery"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type LifecycleDeliveryStateRuntime struct {
	coordinator *platformtx.Coordinator
	binding     lifecycledeliverypostgres.Binding
}

func NewLifecycleDeliveryStateRuntime(
	coordinator *platformtx.Coordinator,
) (*LifecycleDeliveryStateRuntime, error) {
	if coordinator == nil {
		return nil, lifecycledeliveryapp.ErrInvalid
	}
	binding, err := lifecycledeliverypostgres.NewBinding()
	if err != nil {
		return nil, err
	}
	return &LifecycleDeliveryStateRuntime{coordinator: coordinator, binding: binding}, nil
}

func (runtime *LifecycleDeliveryStateRuntime) inTenant(ctx context.Context,
	tenantID uuid.UUID,
	operation func(context.Context, platformtx.JobTenantScope,
		lifecycledeliveryapp.Store) error) error {
	if runtime == nil || runtime.coordinator == nil || ctx == nil ||
		tenantID == uuid.Nil || operation == nil {
		return lifecycledeliveryapp.ErrInvalid
	}
	return runtime.coordinator.InJobTenant(ctx, tenantID,
		func(scoped context.Context, scope platformtx.JobTenantScope) error {
			if !platformtx.ContextOwnsJobScope(scoped, scope) ||
				platformtx.NetworkAllowed(scoped) {
				return lifecycledeliveryapp.ErrInvalid
			}
			store, err := runtime.binding.Bind(scope)
			if err != nil || store == nil {
				return errors.Join(lifecycledeliveryapp.ErrInvalid, err)
			}
			return operation(scoped, scope, store)
		})
}
