package worker

import (
	"fmt"
	"log/slog"
	"net/http"

	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpserver"
	"github.com/Onurlulardan/harpy/core/internal/platform/management"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

type workerRuntime struct {
	poller           *health.Poller
	management       *http.Server
	telemetryRuntime *telemetry.Runtime
	asyncWorkload    *asyncSupervisor
	iamResources     *iamWorkerResources
	deploymentPolicy *tenancymodule.DeploymentPolicyRuntime
	admission        *admission.Gate
}

func buildWorkerRuntime(
	settings config.Worker,
	info buildinfo.Info,
	pool *postgres.Pool,
	monitor *health.Monitor,
	asyncWorkload *asyncSupervisor,
	iamResources *iamWorkerResources,
	policyRuntime *tenancymodule.DeploymentPolicyRuntime,
	workerAdmission *admission.Gate,
	logger *slog.Logger,
	telemetryRuntime *telemetry.Runtime,
) (workerRuntime, error) {
	if pool == nil || monitor == nil || asyncWorkload == nil ||
		iamResources == nil || policyRuntime == nil || workerAdmission == nil ||
		logger == nil ||
		telemetryRuntime == nil {
		return workerRuntime{}, fmt.Errorf("worker runtime dependencies are incomplete")
	}
	report := reportRuntimeError(logger)
	managementHandler, err := management.NewHandler(monitor, info, report)
	if err != nil {
		return workerRuntime{}, err
	}
	errorLog, err := telemetry.StandardErrorLog(logger)
	if err != nil {
		return workerRuntime{}, err
	}
	managementServer, err := httpserver.NewManagement(
		settings.LongRunning().ManagementListenAddress(), managementHandler, errorLog,
	)
	if err != nil {
		return workerRuntime{}, err
	}
	runtimeSettings := settings.LongRunning().Runtime()
	poller, err := health.NewPoller(
		monitor, runtimeSettings.Postgres().HealthCheckPeriod(),
		runtimeSettings.ReadinessTimeout(),
		probeReadiness(pool, policyRuntime.Ready),
		observeWorkerReadiness(workerAdmission),
	)
	if err != nil {
		return workerRuntime{}, err
	}
	return workerRuntime{
		poller: poller, management: managementServer,
		telemetryRuntime: telemetryRuntime, asyncWorkload: asyncWorkload,
		iamResources: iamResources, deploymentPolicy: policyRuntime,
		admission: workerAdmission,
	}, nil
}
