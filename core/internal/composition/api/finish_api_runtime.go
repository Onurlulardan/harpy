package api

import (
	"fmt"
	"log/slog"

	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

func finishAPIRuntime(
	settings config.API, monitor *health.Monitor, info buildinfo.Info,
	logger *slog.Logger, report httpedge.ErrorSink, pool *postgres.Pool,
	policyRuntime *tenancymodule.DeploymentPolicyRuntime,
	telemetryRuntime *telemetry.Runtime, writeCatalogs platformasync.WriteCatalogs,
	iamRuntime iamAPIRuntime, platformOpsHTTP *platformOpsHTTPRuntime,
	billingGroups [][]httpedge.Route, billingLicense []httpedge.Route,
	domains apiDomainRuntimes, deployment deploymentHTTPFoundation,
) (apiRuntime, error) {
	groups := joinAPIRouteGroups(
		[][]httpedge.Route{deployment.routes}, iamRuntime.routes,
		[][]httpedge.Route{platformOpsHTTP.Routes(), billingLicense},
		billingGroups,
	)
	groups = joinAPIRouteGroups(
		groups,
		[][]httpedge.Route{domains.analytics.Routes(), domains.catalog.Routes(),
			domains.documents.Routes(), domains.notifications.Routes(),
			domains.work.Routes()},
		domains.secondary.RouteGroups(),
		[][]httpedge.Route{domains.systemLayouts.Routes(),
			domains.customerProvisioning.Routes()},
	)
	routes, err := buildPublicRoutes(iamRuntime.security, report, groups...)
	if err != nil {
		return apiRuntime{}, err
	}
	restrictedRoutes, err := iamRuntime.postRestore.Wrap(routes)
	if err != nil {
		return apiRuntime{}, fmt.Errorf("build IAM post-restore admission: %w", err)
	}
	return bindFinishedAPIRuntime(
		settings, monitor, info, logger, report, pool, policyRuntime, telemetryRuntime,
		writeCatalogs, iamRuntime, platformOpsHTTP, domains, restrictedRoutes,
	)
}
