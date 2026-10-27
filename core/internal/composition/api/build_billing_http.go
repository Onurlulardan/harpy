package api

import (
	"fmt"

	billingmodule "github.com/Onurlulardan/harpy/core/internal/modules/billing"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func buildBillingHTTP(
	settings config.API,
	pool *postgres.Pool,
	catalogs platformasync.WriteCatalogs,
	iamAccess billingIAMAccess,
	report httpedge.ErrorSink,
) (*billingmodule.Handler, error) {
	if pool == nil || iamAccess == nil || report == nil {
		return nil, fmt.Errorf("Billing HTTP composition dependencies are required")
	}
	coordinator, err := platformtx.NewCoordinator(pool)
	if err != nil {
		return nil, fmt.Errorf("build Billing scope coordinator: %w", err)
	}
	keyring, err := settings.RestQueryCursorKeyring()
	if err != nil {
		return nil, fmt.Errorf("load Billing REST Query cursor keys: %w", err)
	}
	access, err := newBillingAccessVerifier(iamAccess)
	if err != nil {
		return nil, err
	}
	global, err := billingmodule.NewGlobalHTTPRuntime(pool, coordinator, keyring)
	if err != nil {
		return nil, fmt.Errorf("build global Billing HTTP runtime: %w", err)
	}
	tenant, err := billingmodule.NewTenantHTTPRuntime(pool, coordinator, catalogs)
	if err != nil {
		return nil, fmt.Errorf("build tenant Billing HTTP runtime: %w", err)
	}
	reports, err := billingmodule.NewPlatformReportsRuntime(
		coordinator, keyring, catalogs,
	)
	if err != nil {
		return nil, fmt.Errorf("build platform Billing reports: %w", err)
	}
	handler, err := billingmodule.New(billingmodule.Dependencies{
		Access: access, Public: global, Tenant: tenant, Catalog: global,
		Reports: reports, Webhook: billingmodule.NewManualOnlyWebhookHTTPIngress(),
		Report: billingmodule.ErrorSink(report),
	})
	if err != nil {
		return nil, fmt.Errorf("build canonical Billing HTTP handler: %w", err)
	}
	return handler, nil
}
