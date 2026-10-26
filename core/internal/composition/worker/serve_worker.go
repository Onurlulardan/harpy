package worker

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

func serveWorker(
	process *lifecycle.Process,
	settings config.Worker,
	pool *postgres.Pool,
	monitor *health.Monitor,
	runtime workerRuntime,
	management *httpserver.Bound,
	logger *slog.Logger,
) processcode.Code {
	defer func() {
		_ = runtime.iamResources.Close()
	}()
	longRunning := settings.LongRunning()
	host, err := processhost.New(processhost.Dependencies{
		Monitor: monitor, Poller: runtime.poller, Management: management,
		CloseAdmission: runtime.admission.BeginDrain, CloseDatabase: pool.Close,
		ForceCloseDatabase: pool.ForceClose,
		ShutdownTelemetry:  runtime.telemetryRuntime.Shutdown,
		Workload: processhost.Workload{
			BeginDrain: runtime.asyncWorkload.BeginDrain,
			Wait:       runtime.asyncWorkload.Wait,
			Force:      runtime.asyncWorkload.Force,
			Failures:   runtime.asyncWorkload.Failures(),
		},
		DrainTimeout:      longRunning.DrainTimeout(),
		ForceFlushTimeout: longRunning.Runtime().Telemetry().ExportTimeout(),
	})
	if err != nil {
		_ = management.Close()
		pool.Close()
		shutdownStartupTelemetry(longRunning.Runtime(), runtime.telemetryRuntime)
		return processcode.InvalidStartup
	}
	if err := runtime.asyncWorkload.Start(process.RuntimeContext()); err != nil {
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
