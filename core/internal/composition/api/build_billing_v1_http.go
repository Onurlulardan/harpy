package api

import (
	"context"
	"fmt"

	billingmodule "github.com/Onurlulardan/harpy/core/internal/modules/billing"
	billingv1 "github.com/Onurlulardan/harpy/core/internal/modules/billing/compat/billingv1"
	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	cutoverpostgres "github.com/Onurlulardan/harpy/core/internal/platform/cutover/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func buildBillingV1HTTP(
	ctx context.Context,
	settings config.API,
	policy tenancymodule.DeploymentPolicy,
	pool *postgres.Pool,
	catalogs platformasync.WriteCatalogs,
	iamAccess billingIAMAccess,
) (*billingv1.Handler, error) {
	if ctx == nil || ctx.Err() != nil || pool == nil || iamAccess == nil {
		return nil, fmt.Errorf("Billing V1 HTTP composition dependencies are required")
	}
	coordinator, err := platformtx.NewCoordinator(pool)
	if err != nil {
		return nil, fmt.Errorf("build Billing V1 scope coordinator: %w", err)
	}
	keyring, err := settings.RestQueryCursorKeyring()
	if err != nil {
		return nil, fmt.Errorf("load Billing V1 REST Query cursor keys: %w", err)
	}
	canonicalAccess, err := newBillingAccessVerifier(iamAccess)
	if err != nil {
		return nil, err
	}
	global, err := billingmodule.NewGlobalHTTPRuntime(pool, coordinator, keyring)
	if err != nil {
		return nil, fmt.Errorf("build global Billing V1 runtime: %w", err)
	}
	tenant, err := billingmodule.NewTenantHTTPRuntime(pool, coordinator, catalogs)
	if err != nil {
		return nil, fmt.Errorf("build tenant Billing V1 runtime: %w", err)
	}
	reports, err := billingmodule.NewPlatformReportsRuntime(coordinator, keyring, catalogs)
	if err != nil {
		return nil, fmt.Errorf("build Billing V1 reports: %w", err)
	}
	retirement, err := postgres.BindAPIStore(
		pool, cutoverpostgres.NewContractRetirementRuntime,
	)
	if err != nil {
		return nil, err
	}
	availability, err := billingV1Availability(policy)
	if err != nil {
		return nil, err
	}
	handler, err := billingv1.NewProduction(billingv1.Dependencies{
		Access: canonicalAccess, Public: global, Tenant: tenant, Catalog: global,
		Identity: global,
		Reports:  reports, Retirement: retirement, Availability: availability,
	})
	if err != nil {
		return nil, fmt.Errorf("build Billing V1 compatibility handler: %w", err)
	}
	if err := handler.Admit(ctx); err != nil {
		return nil, err
	}
	return handler, nil
}

func billingV1Availability(
	policy tenancymodule.DeploymentPolicy,
) (billingv1.AvailabilityPolicy, error) {
	if !policy.IsValid() {
		return nil, fmt.Errorf("Billing deployment policy is invalid")
	}
	switch policy.Mode() {
	case tenancymodule.DeploymentModeSaaS:
		return billingv1.NewAvailabilityPolicy(billingv1.DeploymentSaaS)
	case tenancymodule.DeploymentModeOnPremise:
		return billingv1.NewAvailabilityPolicy(billingv1.DeploymentOnPrem)
	default:
		return nil, fmt.Errorf("Billing deployment mode is invalid")
	}
}
