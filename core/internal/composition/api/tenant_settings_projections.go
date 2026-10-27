package api

import (
	"context"

	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type tenantSettingsProjectionReader struct{}

func (tenantSettingsProjectionReader) Read(
	_ context.Context, _ platformtx.TenantScope,
) (settings.Projection, error) {
	return settings.Projection{}, nil
}

var _ tenancymodule.TenantSettingsProjectionReader = tenantSettingsProjectionReader{}
