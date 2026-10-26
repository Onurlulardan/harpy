package worker

import (
	"log/slog"
	"net"

	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpserver"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

func bindWorker(
	process *lifecycle.Process,
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
) processcode.Code {
	runtime, err := buildWorkerRuntime(
		settings, info, pool, monitor, asyncWorkload, iamResources,
		policyRuntime, workerAdmission, logger, telemetryRuntime,
	)
	if err != nil {
		reportFailure(nil, logger, "build_contract")
		if iamResources != nil {
			_ = iamResources.Close()
		}
		pool.Close()
		shutdownStartupTelemetry(settings.LongRunning().Runtime(), telemetryRuntime)
		return processcode.InvalidStartup
	}
	management, err := httpserver.Bind(
		process.StartupContext(), runtime.management, net.ListenConfig{},
	)
	if err != nil {
		reportFailure(nil, logger, "runtime")
		_ = runtime.iamResources.Close()
		pool.Close()
		shutdownStartupTelemetry(settings.LongRunning().Runtime(), telemetryRuntime)
		return processcode.ServeFailure
	}
	return serveWorker(process, settings, pool, monitor, runtime, management, logger)
}
