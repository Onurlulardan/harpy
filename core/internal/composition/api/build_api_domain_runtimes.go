package api

import (
	"context"

	documentsmodule "github.com/Onurlulardan/harpy/core/internal/modules/documents"
	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type apiDomainRuntimes struct {
	analytics            analyticsHTTPRuntime
	catalog              catalogHTTPRuntime
	documents            *documentsmodule.HTTPHandler
	documentsProduction  *documentsmodule.Production
	notifications        notificationsHTTPRuntime
	work                 workHTTPRuntime
	secondary            secondaryHTTPRuntime
	systemLayouts        *systemResourceLayoutHTTPHandler
	customerProvisioning *customerProvisioningRuntime
}

func buildAPIDomainRuntimes(
	ctx context.Context, settings config.API, pool *postgres.Pool,
	catalogs platformasync.WriteCatalogs, iam *iamAPIRuntime,
	policyRuntime *tenancymodule.DeploymentPolicyRuntime,
	report httpedge.ErrorSink, info buildinfo.Info, environmentID string,
) (apiDomainRuntimes, error) {
	analyticsHTTP, err := buildAnalyticsHTTP(
		ctx, settings, pool, catalogs, *iam, report,
	)
	if err != nil {
		return apiDomainRuntimes{}, err
	}
	catalogHTTP, err := buildCatalogHTTP(
		settings, pool, catalogs, *iam, report,
	)
	if err != nil {
		return apiDomainRuntimes{}, err
	}
	documentsHTTP, documentsProduction, err := buildDocumentsHTTP(
		ctx, settings, pool, catalogs, *iam, report, info,
	)
	if err != nil {
		return apiDomainRuntimes{}, err
	}
	notificationsHTTP, err := buildNotificationsHTTP(
		settings, pool, catalogs, iam.notificationsAccess, report)
	if err != nil {
		return apiDomainRuntimes{}, err
	}
	workHTTP, err := buildWorkHTTP(settings, pool, catalogs, iam.workAccess, report)
	if err != nil {
		return apiDomainRuntimes{}, err
	}
	secondaryHTTP, err := buildSecondaryHTTPRuntime(
		settings, pool, iam.coordinator, iam, report, environmentID,
	)
	if err != nil {
		return apiDomainRuntimes{}, err
	}
	systemLayoutHTTP, err := buildSystemResourceLayoutHTTP(
		iam.systemLayouts, iam.systemLayoutAccess, report,
	)
	if err != nil {
		return apiDomainRuntimes{}, err
	}
	customerProvisioning, err := buildCustomerProvisioningRuntime(ctx, settings,
		pool, policyRuntime, iam.customerProvisioningAccess, iam.deliverySealer, report)
	if err != nil {
		return apiDomainRuntimes{}, err
	}
	return apiDomainRuntimes{
		analytics: analyticsHTTP, catalog: catalogHTTP,
		documents: documentsHTTP, documentsProduction: documentsProduction,
		notifications: notificationsHTTP, work: workHTTP, secondary: secondaryHTTP,
		systemLayouts: systemLayoutHTTP, customerProvisioning: customerProvisioning,
	}, nil
}
