package tenancy

import (
	"context"
	"fmt"

	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type TenantSettingsAuthority struct {
	TenancyGroup, Tenancy               int64
	IAMGroup, IAM, SystemResourceLayout int64
	NotificationsGroup, Notifications   int64
	Billing, Forms                      int64
}

type TenantSettingsGenerationResolver interface {
	ResolveTenantSettingsAuthority(context.Context, uuid.UUID) (TenantSettingsAuthority, error)
}

type TenantSettingsScopeCoordinator interface {
	InTenant(context.Context, platformtx.ScopeRequest,
		func(context.Context, platformtx.TenantScope) error) error
}

// TenantSettingsSnapshotAuthority is implemented by the single platform
// snapshot-admission owner. Tenancy declares the exact authority set but does
// not mint a snapshot capability itself.
type TenantSettingsSnapshotAuthority interface {
	Issue(string, map[string]int64) (platformtx.VerifiedOperation, error)
}

type tenantSettingsApplication interface {
	Get(context.Context, platformtx.TenantScope) (settings.Snapshot, error)
	Update(context.Context, platformtx.TenantScope, settings.UpdateCommand) (settings.UpdateResult, error)
	TestEmail(context.Context, platformtx.TenantScope, settings.TestEmailCommand) (settings.TestEmailResult, error)
}

type TenantSettingsHTTPRuntime struct {
	resolver      TenantSettingsGenerationResolver
	coordinator   TenantSettingsScopeCoordinator
	application   tenantSettingsApplication
	admitCommand  tenantSettingsCommandAdmitter
	issueSnapshot tenantSettingsSnapshotIssuer
}

type tenantSettingsCommandAdmitter func(
	uuid.UUID, uuid.UUID, string,
	[]platformtx.CommandAuthorityGroup,
	[]platformtx.CommandAuthorityDomain,
	[]string,
) (platformtx.ScopeRequest, error)

type tenantSettingsSnapshotIssuer func(
	string, map[string]int64,
) (platformtx.VerifiedOperation, error)

func NewTenantSettingsHTTPRuntime(resolver TenantSettingsGenerationResolver,
	coordinator TenantSettingsScopeCoordinator, application tenantSettingsApplication,
	snapshotAuthority TenantSettingsSnapshotAuthority,
) (*TenantSettingsHTTPRuntime, error) {
	if tenantSettingsPortMissing(resolver) || tenantSettingsPortMissing(coordinator) ||
		tenantSettingsPortMissing(application) || tenantSettingsPortMissing(snapshotAuthority) {
		return nil, fmt.Errorf("tenant settings HTTP runtime dependencies are required")
	}
	return &TenantSettingsHTTPRuntime{resolver: resolver, coordinator: coordinator,
		application: application, admitCommand: platformtx.AdmitMultiGroupCommandScope,
		issueSnapshot: snapshotAuthority.Issue}, nil
}

func validTenantSettingsSnapshotAuthority(value TenantSettingsAuthority) bool {
	return value.Tenancy > 0 && value.IAM > 0 && value.Notifications > 0 &&
		value.Billing > 0
}

func validTenantSettingsCommandAuthority(value TenantSettingsAuthority,
	includeNotifications bool,
) bool {
	valid := value.TenancyGroup > 0 && value.Tenancy > 0 &&
		value.IAMGroup > 0 && value.IAM > 0 && value.SystemResourceLayout > 0
	return valid && (!includeNotifications ||
		value.NotificationsGroup > 0 && value.Notifications > 0)
}
