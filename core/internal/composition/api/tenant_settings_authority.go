package api

import (
	"context"
	"fmt"

	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

type tenantSettingsGenerationResolver struct {
	pool *postgres.Pool
}

func newTenantSettingsGenerationResolver(
	pool *postgres.Pool,
) (*tenantSettingsGenerationResolver, error) {
	if pool == nil {
		return nil, fmt.Errorf("tenant settings generation pool is required")
	}
	return &tenantSettingsGenerationResolver{pool: pool}, nil
}

func (resolver *tenantSettingsGenerationResolver) ResolveTenantSettingsAuthority(
	ctx context.Context, tenantID uuid.UUID,
) (tenancymodule.TenantSettingsAuthority, error) {
	if resolver == nil || resolver.pool == nil {
		return tenancymodule.TenantSettingsAuthority{},
			tenancymodule.ErrTenantSettingsUnavailable
	}
	iamGroup, iam, layout, err := resolver.pool.ResolveTenantIAMGenerations(ctx, tenantID)
	if err != nil {
		return tenancymodule.TenantSettingsAuthority{}, err
	}
	notificationGroup, notifications, err := resolver.pool.ResolveTenantNotificationGenerations(
		ctx, tenantID)
	if err != nil {
		return tenancymodule.TenantSettingsAuthority{}, err
	}
	_, billing, err := resolver.pool.ResolveTenantBillingGenerations(ctx, tenantID)
	if err != nil {
		return tenancymodule.TenantSettingsAuthority{}, err
	}
	tenancyGroup, tenancy, err := resolver.pool.ResolveTenantTenancyGenerations(ctx, tenantID)
	if err != nil {
		return tenancymodule.TenantSettingsAuthority{}, err
	}
	return tenancymodule.TenantSettingsAuthority{
		TenancyGroup: tenancyGroup, Tenancy: tenancy,
		IAMGroup: iamGroup, IAM: iam, SystemResourceLayout: layout,
		NotificationsGroup: notificationGroup, Notifications: notifications,
		Billing: billing,
	}, nil
}

var _ tenancymodule.TenantSettingsGenerationResolver = (*tenantSettingsGenerationResolver)(nil)
