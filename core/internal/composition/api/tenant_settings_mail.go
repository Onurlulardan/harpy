package api

import (
	"context"

	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type tenantSettingsMailUnavailable struct{}

func (tenantSettingsMailUnavailable) Replace(
	_ context.Context, _ platformtx.TenantScope, _ settings.MailMutation,
) error {
	return settings.ErrUnavailable
}

func (tenantSettingsMailUnavailable) EnqueueTest(
	_ context.Context, _ platformtx.TenantScope, _ settings.TestMailRequest,
) (settings.TestEmailResult, error) {
	return settings.TestEmailResult{}, settings.ErrUnavailable
}

var _ tenancymodule.TenantSettingsMail = tenantSettingsMailUnavailable{}
