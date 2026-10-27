package api

import (
	"context"

	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func buildAPIBilling(
	ctx context.Context, settings config.API, pool *postgres.Pool,
	catalogs platformasync.WriteCatalogs, policyRuntime *tenancymodule.DeploymentPolicyRuntime,
	deployment deploymentHTTPFoundation, iamRuntime iamAPIRuntime, report httpedge.ErrorSink,
) ([][]httpedge.Route, []httpedge.Route, error) {
	billingHTTP, err := buildBillingHTTP(
		settings, pool, catalogs, iamRuntime.billingAccess, report,
	)
	if err != nil {
		return nil, nil, err
	}
	billingLicenseHTTP, err := buildBillingLicenseHTTP(
		pool, policyRuntime, iamRuntime.platformAccess, report)
	if err != nil {
		return nil, nil, err
	}
	billingV1HTTP, err := buildBillingV1HTTP(
		ctx, settings, deployment.policy, pool, catalogs, iamRuntime.billingAccess,
	)
	if err != nil {
		return nil, nil, err
	}
	billingGroups, err := buildBillingRouteGroups(billingHTTP, billingV1HTTP)
	if err != nil {
		return nil, nil, err
	}
	return billingGroups, billingLicenseHTTP.Routes(), nil
}
