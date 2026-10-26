package api

import (
	"log/slog"
	"net/http"

	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

func bindFinishedAPIRuntime(
	settings config.API, monitor *health.Monitor, info buildinfo.Info,
	logger *slog.Logger, report httpedge.ErrorSink, pool *postgres.Pool,
	policyRuntime *tenancymodule.DeploymentPolicyRuntime,
	telemetryRuntime *telemetry.Runtime, writeCatalogs platformasync.WriteCatalogs,
	iamRuntime iamAPIRuntime, platformOpsHTTP *platformOpsHTTPRuntime,
	domains apiDomainRuntimes, restrictedRoutes http.Handler,
) (apiRuntime, error) {
	admission := &httpedge.Admission{}
	serving, err := buildAPIServingRuntime(
		settings, monitor, info, logger, admission, restrictedRoutes, report,
		platformOpsHTTP.Observe, domains.secondary.Observe, pool, policyRuntime,
		apiRuntimeReadiness(
			iamRuntime, domains.analytics, domains.catalog,
			domains.documents, domains.notifications, domains.work, domains.secondary,
			domains.systemLayouts,
		),
	)
	if err != nil {
		return apiRuntime{}, err
	}
	return apiRuntime{
		admission: admission, poller: serving.poller,
		management: serving.management, public: serving.public,
		telemetryRuntime: telemetryRuntime, asyncWriteCatalogs: writeCatalogs,
		analytics: domains.analytics,
		deploymentPolicy:     policyRuntime,
		catalog:              domains.catalog,
		documents:            domains.documents,
		documentsProduction:  domains.documentsProduction,
		notifications:        domains.notifications,
		work:                 domains.work,
		secondary:            domains.secondary,
		systemLayouts:        domains.systemLayouts,
		customerProvisioning: domains.customerProvisioning,
		platformOps:          platformOpsHTTP,
		iam:                  iamRuntime,
	}, nil
}
