package api

import (
	"context"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

func startAPI(
	process *lifecycle.Process,
	settings config.API,
	info buildinfo.Info,
	logger *slog.Logger,
	telemetryRuntime *telemetry.Runtime,
) processcode.Code {
	return startAPIWithOpen(
		process, settings, info, logger, telemetryRuntime, postgres.Open,
	)
}

type apiPoolOpener func(
	context.Context,
	config.Postgres,
	postgres.Profile,
) (*postgres.Pool, error)

func startAPIWithOpen(
	process *lifecycle.Process,
	settings config.API,
	info buildinfo.Info,
	logger *slog.Logger,
	telemetryRuntime *telemetry.Runtime,
	openPool apiPoolOpener,
) processcode.Code {
	runtimeSettings := settings.LongRunning().Runtime()
	startupContext, cancelStartup := context.WithTimeout(
		process.StartupContext(), runtimeSettings.ReadinessTimeout(),
	)
	pool, err := openPool(
		startupContext, runtimeSettings.Postgres(), postgres.ProfileAPI,
	)
	if err != nil {
		cancelStartup()
		reportFailure(nil, logger, "dependency_preflight")
		shutdownStartupTelemetry(runtimeSettings, telemetryRuntime)
		return processcode.DependencyPreflight
	}
	expectedMode, err := settings.DeploymentMode()
	if err != nil {
		cancelStartup()
		pool.Close()
		shutdownStartupTelemetry(runtimeSettings, telemetryRuntime)
		return processcode.InvalidStartup
	}
	policyRuntime, err := bindDeploymentPolicy(pool, expectedMode)
	if err != nil {
		cancelStartup()
		reportFailure(nil, logger, "dependency_preflight")
		pool.Close()
		shutdownStartupTelemetry(runtimeSettings, telemetryRuntime)
		return processcode.DependencyPreflight
	}
	initial := probeReadiness(pool, policyRuntime.Ready)(startupContext)
	cancelStartup()
	if !allReady(initial) {
		reportFailureCause(nil, logger, "dependency_preflight", readinessFailure(initial))
		pool.Close()
		shutdownStartupTelemetry(runtimeSettings, telemetryRuntime)
		return processcode.DependencyPreflight
	}
	monitor, err := health.NewMonitor(
		info.Service(), runtimeSettings.Postgres().HealthCheckPeriod(), nil,
	)
	if err != nil {
		pool.Close()
		shutdownStartupTelemetry(runtimeSettings, telemetryRuntime)
		return processcode.InvalidStartup
	}
	return bindAPI(
		process, settings, info, pool, policyRuntime, monitor, logger,
		telemetryRuntime,
	)
}
