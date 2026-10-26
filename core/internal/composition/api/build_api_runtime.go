package api

import (
	"context"
	"fmt"
	"log/slog"

	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

func buildAPIRuntime(
	ctx context.Context, settings config.API,
	info buildinfo.Info, pool *postgres.Pool,
	policyRuntime *tenancymodule.DeploymentPolicyRuntime,
	monitor *health.Monitor,
	logger *slog.Logger,
	telemetryRuntime *telemetry.Runtime,
) (apiRuntime, error) {
	if ctx == nil || ctx.Err() != nil || pool == nil || policyRuntime == nil ||
		monitor == nil || logger == nil || telemetryRuntime == nil {
		return apiRuntime{}, fmt.Errorf("API runtime dependencies are incomplete")
	}
	writeCatalogs, err := platformasync.LoadWriteCatalogs()
	if err != nil {
		return apiRuntime{}, fmt.Errorf("build async write catalogs: %w", err)
	}
	report := reportRuntimeError(logger)
	deployment, err := buildDeploymentHTTPFoundation(
		info, policyRuntime, report,
	)
	if err != nil {
		return apiRuntime{}, err
	}
	iamRuntime, err := buildDeploymentIAMAPIRuntime(
		ctx, settings, pool, writeCatalogs, report, deployment.policy,
	)
	if err != nil {
		return apiRuntime{}, err
	}
	iamOwned := true
	defer func() {
		if iamOwned {
			iamRuntime.Close()
		}
	}()
	runtime, err := assembleAPIRuntime(
		ctx, settings, info, pool, policyRuntime, monitor, logger, telemetryRuntime,
		writeCatalogs, report, deployment, iamRuntime,
	)
	if err != nil {
		return apiRuntime{}, err
	}
	iamOwned = false
	return runtime, nil
}
