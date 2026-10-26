package iam

import (
	"context"
	"errors"

	invitationdeliverypostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/invitationdelivery"
	invitationdeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitationdelivery"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type InvitationDeliveryRuntime struct {
	coordinator *platformtx.Coordinator
	binding     invitationdeliverypostgres.Binding
}

func NewInvitationDeliveryRuntime(
	coordinator *platformtx.Coordinator,
) (*InvitationDeliveryRuntime, error) {
	if coordinator == nil {
		return nil, invitationdeliveryapp.ErrInvalid
	}
	binding, err := invitationdeliverypostgres.NewBinding()
	if err != nil {
		return nil, err
	}
	return &InvitationDeliveryRuntime{
		coordinator: coordinator, binding: binding,
	}, nil
}

func (runtime *InvitationDeliveryRuntime) Claim(
	ctx context.Context,
	command invitationdeliveryapp.ClaimCommand,
) (invitationdeliveryapp.ClaimDecision, string, error) {
	var result invitationdeliveryapp.ClaimDecision
	var tenantSlug string
	err := runtime.inTenant(ctx, command.TenantID, func(
		scoped context.Context,
		scope platformtx.JobTenantScope,
		store invitationdeliveryapp.Store,
	) error {
		service, err := invitationdeliveryapp.NewService(store)
		if err != nil {
			return err
		}
		result, err = service.Claim(scoped, command)
		tenantSlug = scope.TenantSlug()
		return err
	})
	return result, tenantSlug, err
}

func (runtime *InvitationDeliveryRuntime) Release(
	ctx context.Context,
	reference invitationdeliveryapp.EffectReference,
) error {
	return runtime.inTenant(ctx, reference.TenantID, func(
		scoped context.Context,
		_ platformtx.JobTenantScope,
		store invitationdeliveryapp.Store,
	) error {
		service, err := invitationdeliveryapp.NewService(store)
		if err != nil {
			return err
		}
		return service.Release(scoped, reference)
	})
}

func (runtime *InvitationDeliveryRuntime) MarkUnknown(
	ctx context.Context,
	reference invitationdeliveryapp.EffectReference,
) error {
	return runtime.inTenant(ctx, reference.TenantID, func(
		scoped context.Context,
		_ platformtx.JobTenantScope,
		store invitationdeliveryapp.Store,
	) error {
		service, err := invitationdeliveryapp.NewService(store)
		if err != nil {
			return err
		}
		return service.MarkUnknown(scoped, reference)
	})
}

func (runtime *InvitationDeliveryRuntime) Finalize(
	ctx context.Context,
	command invitationdeliveryapp.FinalizeCommand,
) (invitationdeliveryapp.StoredResult, error) {
	var result invitationdeliveryapp.StoredResult
	err := runtime.inTenant(ctx, command.TenantID, func(
		scoped context.Context,
		_ platformtx.JobTenantScope,
		store invitationdeliveryapp.Store,
	) error {
		service, err := invitationdeliveryapp.NewService(store)
		if err != nil {
			return err
		}
		result, err = service.Finalize(scoped, command)
		return err
	})
	return result, err
}

func (runtime *InvitationDeliveryRuntime) inTenant(
	ctx context.Context,
	tenantID uuid.UUID,
	operation func(context.Context, platformtx.JobTenantScope,
		invitationdeliveryapp.Store) error,
) error {
	if runtime == nil || runtime.coordinator == nil || ctx == nil ||
		tenantID == uuid.Nil || operation == nil {
		return invitationdeliveryapp.ErrInvalid
	}
	return runtime.coordinator.InJobTenant(
		ctx, tenantID, func(scoped context.Context, scope platformtx.JobTenantScope) error {
			if !platformtx.ContextOwnsJobScope(scoped, scope) ||
				platformtx.NetworkAllowed(scoped) {
				return invitationdeliveryapp.ErrInvalid
			}
			store, err := runtime.binding.Bind(scope)
			if err != nil || store == nil {
				return errors.Join(invitationdeliveryapp.ErrInvalid, err)
			}
			return operation(scoped, scope, store)
		},
	)
}
