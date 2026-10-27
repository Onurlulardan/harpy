package api

import (
	"log/slog"
	"net"

	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpserver"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

func bindAPI(
	process *lifecycle.Process,
	settings config.API,
	info buildinfo.Info,
	pool *postgres.Pool,
	policyRuntime *tenancymodule.DeploymentPolicyRuntime,
	monitor *health.Monitor,
	logger *slog.Logger,
	telemetryRuntime *telemetry.Runtime,
) processcode.Code {
	runtime, err := buildAPIRuntime(
		process.StartupContext(), settings, info, pool, policyRuntime, monitor,
		logger, telemetryRuntime,
	)
	if err != nil {
		if settings.LongRunning().Runtime().Environment() == config.EnvironmentDevelopment {
			logger.Error(
				"core API composition failed",
				slog.String("error_class", "build_contract"),
				slog.String("service", "core-api"),
				slog.String("cause", err.Error()),
			)
		} else {
			reportFailure(nil, logger, "build_contract")
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
		runtime.Close()
		pool.Close()
		shutdownStartupTelemetry(settings.LongRunning().Runtime(), telemetryRuntime)
		return processcode.ServeFailure
	}
	public, err := httpserver.Bind(
		process.StartupContext(), runtime.public, net.ListenConfig{},
	)
	if err != nil {
		reportFailure(nil, logger, "runtime")
		_ = management.Close()
		runtime.Close()
		pool.Close()
		shutdownStartupTelemetry(settings.LongRunning().Runtime(), telemetryRuntime)
		return processcode.ServeFailure
	}
	return serveAPI(process, settings, pool, monitor, runtime, management, public, logger)
}
