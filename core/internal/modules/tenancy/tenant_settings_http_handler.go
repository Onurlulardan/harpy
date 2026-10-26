package tenancy

import (
	"context"
	"fmt"

	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
)

type tenantSettingsHTTPApplication interface {
	Get(context.Context, TenantSettingsAccess) (settings.Snapshot, error)
	Update(context.Context, TenantSettingsAccess, settings.UpdateCommand) (settings.UpdateResult, error)
	TestEmail(context.Context, TenantSettingsAccess, settings.TestEmailCommand) (settings.TestEmailResult, error)
}

type TenantSettingsHTTP struct {
	access      TenantSettingsAccessVerifier
	application tenantSettingsHTTPApplication
	report      func(error)
}

func NewTenantSettingsHTTP(access TenantSettingsAccessVerifier,
	application tenantSettingsHTTPApplication, report func(error),
) (*TenantSettingsHTTP, error) {
	if tenantSettingsPortMissing(access) || tenantSettingsPortMissing(application) || report == nil {
		return nil, fmt.Errorf("tenant settings HTTP dependencies are required")
	}
	return &TenantSettingsHTTP{access: access, application: application, report: report}, nil
}

func (handler *TenantSettingsHTTP) Ready() bool {
	return handler != nil && !tenantSettingsPortMissing(handler.access) &&
		!tenantSettingsPortMissing(handler.application) && handler.report != nil
}
