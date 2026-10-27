package api

import (
	"context"
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

func assembleAPIRuntime(
	ctx context.Context, settings config.API, info buildinfo.Info, pool *postgres.Pool,
	policyRuntime *tenancymodule.DeploymentPolicyRuntime, monitor *health.Monitor,
	logger *slog.Logger, telemetryRuntime *telemetry.Runtime,
	writeCatalogs platformasync.WriteCatalogs, report httpedge.ErrorSink,
	deployment deploymentHTTPFoundation, iamRuntime iamAPIRuntime,
) (apiRuntime, error) {
	platformOpsHTTP, err := buildPlatformOpsHTTP(
		settings, pool, policyRuntime, monitor, iamRuntime.platformAccess, report,
	)
	if err != nil {
		return apiRuntime{}, err
	}
	billingGroups, billingLicense, err := buildAPIBilling(
		ctx, settings, pool, writeCatalogs, policyRuntime, deployment, iamRuntime, report,
	)
	if err != nil {
		return apiRuntime{}, err
	}
	domains, err := buildAPIDomainRuntimes(
		ctx, settings, pool, writeCatalogs, &iamRuntime, policyRuntime, report, info,
		deployment.policy.EnvironmentID(),
	)
	if err != nil {
		return apiRuntime{}, err
	}
	if err := iamRuntime.buildActorBridge(settings); err != nil {
		return apiRuntime{}, err
	}
	return finishAPIRuntime(
		settings, monitor, info, logger, report, pool, policyRuntime, telemetryRuntime,
		writeCatalogs, iamRuntime, platformOpsHTTP, billingGroups, billingLicense,
		domains, deployment,
	)
}
