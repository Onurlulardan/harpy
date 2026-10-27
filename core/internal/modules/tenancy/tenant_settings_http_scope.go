package tenancy

import (
	"context"

	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
	"github.com/Onurlulardan/harpy/core/internal/platform/actorcontext"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *TenantSettingsHTTPRuntime) snapshot(ctx context.Context,
	access TenantSettingsAccess, operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	authority, err := runtime.authority(ctx, access, TenantSettingsPermissionRead, operation)
	if err != nil || !validTenantSettingsSnapshotAuthority(authority) {
		if err == nil {
			err = settings.ErrUnavailable
		}
		return err
	}
	actor, err := actorcontext.Issue(access.principalID, access.tenantID)
	if err != nil {
		return settings.ErrUnavailable
	}
	if runtime.issueSnapshot == nil {
		return settings.ErrUnavailable
	}
	verified, err := runtime.issueSnapshot(operation,
		map[string]int64{"billing": authority.Billing, "iam": authority.IAM,
			"notifications": authority.Notifications, "tenancy": authority.Tenancy})
	if err != nil {
		return settings.ErrUnavailable
	}
	return runtime.coordinator.InTenant(ctx, platformtx.ScopeRequest{
		Actor: actor, Operation: verified, Mode: platformtx.SnapshotQuery}, action)
}

func (runtime *TenantSettingsHTTPRuntime) command(ctx context.Context,
	access TenantSettingsAccess, operation string, includeNotifications bool,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	authority, err := runtime.authority(ctx, access, TenantSettingsPermissionManage, operation)
	if err != nil || !validTenantSettingsCommandAuthority(authority, includeNotifications) {
		if err == nil {
			err = settings.ErrUnavailable
		}
		return err
	}
	if runtime.admitCommand == nil {
		return settings.ErrUnavailable
	}
	groups := []platformtx.CommandAuthorityGroup{
		{ID: "iam_system_resources", Generation: authority.IAMGroup},
		{ID: "tenancy", Generation: authority.TenancyGroup},
	}
	domains := []platformtx.CommandAuthorityDomain{{Name: "iam", Generation: authority.IAM},
		{Name: "system_resource_layout", Generation: authority.SystemResourceLayout},
		{Name: "tenancy", Generation: authority.Tenancy}}
	participants := []string{"tenancy"}
	if includeNotifications {
		groups = append(groups, platformtx.CommandAuthorityGroup{ID: "notifications",
			Generation: authority.NotificationsGroup})
		domains = append(domains, platformtx.CommandAuthorityDomain{Name: "notifications",
			Generation: authority.Notifications})
		participants = append(participants, "notifications")
	}
	request, err := runtime.admitCommand(access.principalID,
		access.tenantID, operation, groups, domains, participants)
	if err != nil {
		return settings.ErrUnavailable
	}
	return runtime.coordinator.InTenant(ctx, request, action)
}

func (runtime *TenantSettingsHTTPRuntime) authority(ctx context.Context,
	access TenantSettingsAccess, permission TenantSettingsPermission, operation string,
) (TenantSettingsAuthority, error) {
	if runtime == nil || ctx == nil || ctx.Err() != nil || !access.validFor(permission, operation) {
		return TenantSettingsAuthority{}, settings.ErrForbidden
	}
	value, err := runtime.resolver.ResolveTenantSettingsAuthority(ctx, access.tenantID)
	if err != nil {
		return TenantSettingsAuthority{}, settings.ErrUnavailable
	}
	return value, nil
}
