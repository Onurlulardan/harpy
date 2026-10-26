package api

import (
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpserver"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/processhost"
)

func serveAPI(
	process *lifecycle.Process,
	settings config.API,
	pool *postgres.Pool,
	monitor *health.Monitor,
	runtime apiRuntime,
	management *httpserver.Bound,
	public *httpserver.Bound,
	logger *slog.Logger,
) processcode.Code {
	defer runtime.Close()
	longRunning := settings.LongRunning()
	host, err := processhost.New(processhost.Dependencies{
		Monitor: monitor, Poller: runtime.poller,
		Management: management, Public: public,
		CloseAdmission:     runtime.admission.BeginDrain,
		CloseDatabase:      pool.Close,
		ForceCloseDatabase: pool.ForceClose,
		ShutdownTelemetry:  runtime.telemetryRuntime.Shutdown,
		Workload: processhost.Workload{
			BeginDrain: runtime.notifications.workload.BeginDrain,
			Wait:       runtime.notifications.workload.Wait,
			Force:      runtime.notifications.workload.Force,
			Failures:   runtime.notifications.workload.Failures(),
		},
		DrainTimeout:      longRunning.DrainTimeout(),
		ForceFlushTimeout: longRunning.Runtime().Telemetry().ExportTimeout(),
	})
	if err != nil {
		_ = public.Close()
		_ = management.Close()
		pool.Close()
		shutdownStartupTelemetry(longRunning.Runtime(), runtime.telemetryRuntime)
		return processcode.InvalidStartup
	}
	if err := runtime.notifications.workload.Start(process.RuntimeContext()); err != nil {
		_ = public.Close()
		_ = management.Close()
		pool.Close()
		shutdownStartupTelemetry(longRunning.Runtime(), runtime.telemetryRuntime)
		return processcode.InvalidStartup
	}
	outcome := host.RunProcess(process)
	if outcome.ExitCode() == processcode.ServeFailure {
		reportFailure(nil, logger, "runtime")
	}
	if outcome.ExitCode() == processcode.DrainForced {
		reportFailure(nil, logger, "drain")
	}
	return outcome.ExitCode()
}
