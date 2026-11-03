package api

import (
	"context"
	"io"
	"os"

	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/envfile"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

func Run(arguments []string) int {
	process := lifecycle.NotifyProcess(context.Background())
	defer process.Stop()
	return runProcess(process, arguments, os.Stderr).Int()
}

func run(
	ctx context.Context,
	arguments []string,
	stderr io.Writer,
) processcode.Code {
	return runProcess(lifecycle.LocalProcess(ctx), arguments, stderr)
}

func runProcess(
	process *lifecycle.Process,
	arguments []string,
	stderr io.Writer,
) processcode.Code {
	ctx := process.StartupContext()
	arguments, err := envfile.Consume(arguments, "api")
	if err != nil {
		reportFailureCause(stderr, nil, "configuration", err.Error())
		return processcode.InvalidStartup
	}
	if err := validateInvocation(arguments); err != nil {
		reportFailureCause(stderr, nil, "invocation", err.Error())
		return processcode.InvalidStartup
	}
	settings, err := config.LoadAPI()
	if err != nil {
		reportFailureCause(stderr, nil, "configuration", err.Error())
		return processcode.InvalidStartup
	}
	info, err := buildinfo.ReadAPI(settings.LongRunning().Runtime().Environment())
	if err != nil {
		reportFailureCause(stderr, nil, "build_contract", err.Error())
		return processcode.InvalidStartup
	}
	logger, err := telemetry.NewLogger(
		stderr, settings.LongRunning().Runtime().LogLevel(), info,
	)
	if err != nil {
		reportFailure(stderr, nil, "configuration")
		return processcode.InvalidStartup
	}
	telemetryRuntime, err := telemetry.New(
		ctx, settings.LongRunning().Runtime().Telemetry(), info, logger,
	)
	if err != nil {
		reportFailure(stderr, logger, "configuration")
		return processcode.InvalidStartup
	}
	return startAPI(process, settings, info, logger, telemetryRuntime)
}
