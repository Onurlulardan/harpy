package tenancy

import (
	settingspostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/tenantsettings"
	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
)

func NewTenantSettingsApplication(projections TenantSettingsProjectionReader,
	mail TenantSettingsMail, clock TenantSettingsClock,
) (*settings.Application, error) {
	registry, err := settingspostgres.New()
	if err != nil {
		return nil, err
	}
	return settings.New(registry, projections, mail, clock)
}
