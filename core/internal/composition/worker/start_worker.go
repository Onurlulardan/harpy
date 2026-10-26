package worker

import (
	"context"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

func startWorker(
	process *lifecycle.Process,
	settings config.Worker,
	info buildinfo.Info,
	embedded schemacontract.EmbeddedContract,
	logger *slog.Logger,
	telemetryRuntime *telemetry.Runtime,
) processcode.Code {
	return startWorkerWithOpen(
		process, settings, info, embedded, logger, telemetryRuntime, postgres.Open,
	)
}

type workerPoolOpener func(
	context.Context,
	config.Postgres,
	postgres.Profile,
	schemacontract.EmbeddedContract,
) (*postgres.Pool, error)

func startWorkerWithOpen(
	process *lifecycle.Process,
	settings config.Worker,
	info buildinfo.Info,
	embedded schemacontract.EmbeddedContract,
	logger *slog.Logger,
	telemetryRuntime *telemetry.Runtime,
	openPool workerPoolOpener,
) processcode.Code {
	runtimeSettings := settings.LongRunning().Runtime()
	startupContext, cancelStartup := context.WithTimeout(
		process.StartupContext(), runtimeSettings.ReadinessTimeout(),
	)
	pool, err := openPool(
		startupContext, runtimeSettings.Postgres(), postgres.ProfileWorker, embedded,
	)
	if err != nil {
		cancelStartup()
		reportFailureCause(nil, logger, "dependency_preflight", err.Error())
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
		reportFailureCause(nil, logger, "dependency_preflight", err.Error())
		pool.Close()
		shutdownStartupTelemetry(runtimeSettings, telemetryRuntime)
		return processcode.DependencyPreflight
	}
	initial := probeReadiness(pool, policyRuntime.Ready)(startupContext)
	if !allReady(initial) {
		cancelStartup()
		reportFailureCause(nil, logger, "dependency_preflight", "initial readiness probe failed")
		pool.Close()
		shutdownStartupTelemetry(runtimeSettings, telemetryRuntime)
		return processcode.DependencyPreflight
	}
	workerAdmission := admission.NewGate(nil)
	if !workerAdmission.Open() {
		cancelStartup()
		reportFailure(nil, logger, "dependency_preflight")
		pool.Close()
		shutdownStartupTelemetry(runtimeSettings, telemetryRuntime)
		return processcode.DependencyPreflight
	}
	cancelStartup()
	retentionObserver, err := telemetryRuntime.NewAsyncRetentionObserver()
	if err != nil {
		reportFailure(nil, logger, "build_contract")
		pool.Close()
		shutdownStartupTelemetry(runtimeSettings, telemetryRuntime)
		return processcode.InvalidStartup
	}
	iamSettings, err := settings.IAM()
	if err != nil {
		reportFailure(nil, logger, "build_contract")
		pool.Close()
		shutdownStartupTelemetry(runtimeSettings, telemetryRuntime)
		return processcode.InvalidStartup
	}
	asyncWorkload, iamResources, err := buildAsyncWorkload(
		process.StartupContext(), pool, settings, settings.RabbitMQ(), iamSettings,
		retentionObserver, workerAdmission, info.Version().Version,
	)
	if err != nil {
		reportFailureCause(nil, logger, "dependency_preflight", err.Error())
		pool.Close()
		shutdownStartupTelemetry(runtimeSettings, telemetryRuntime)
		return processcode.DependencyPreflight
	}
	if err := buildAsyncObservability(pool, telemetryRuntime, logger); err != nil {
		reportFailure(nil, logger, "build_contract")
		_ = iamResources.Close()
		pool.Close()
		shutdownStartupTelemetry(runtimeSettings, telemetryRuntime)
		return processcode.InvalidStartup
	}
	monitor, err := health.NewMonitor(
		info.Service(), runtimeSettings.Postgres().HealthCheckPeriod(), nil,
	)
	if err != nil {
		_ = iamResources.Close()
		pool.Close()
		shutdownStartupTelemetry(runtimeSettings, telemetryRuntime)
		return processcode.InvalidStartup
	}
	return bindWorker(
		process, settings, info, pool, monitor, asyncWorkload, iamResources,
		policyRuntime, workerAdmission, logger, telemetryRuntime,
	)
}
