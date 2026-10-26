package worker

import (
	"context"
	"io"
	"os"
	"runtime/debug"

	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

type schemaContractLoader func() (schemacontract.EmbeddedContract, error)

func Run(arguments []string) int {
	process := lifecycle.NotifyProcess(context.Background())
	defer process.Stop()
	return runProcessWithContractLoader(
		process, arguments, os.Stderr, loadSchemaContract,
	).Int()
}

func run(
	ctx context.Context,
	arguments []string,
	stderr io.Writer,
) processcode.Code {
	return runProcessWithContractLoader(
		lifecycle.LocalProcess(ctx), arguments, stderr, loadSchemaContract,
	)
}

func runWithContractLoader(
	ctx context.Context,
	arguments []string,
	stderr io.Writer,
	loadContract schemaContractLoader,
) processcode.Code {
	return runProcessWithContractLoader(
		lifecycle.LocalProcess(ctx), arguments, stderr, loadContract,
	)
}

func runProcessWithContractLoader(
	process *lifecycle.Process,
	arguments []string,
	stderr io.Writer,
	loadContract schemaContractLoader,
) processcode.Code {
	ctx := process.StartupContext()
	if err := validateInvocation(arguments); err != nil {
		reportFailure(stderr, nil, "invocation")
		return processcode.InvalidStartup
	}
	settings, err := config.LoadWorker()
	if err != nil {
		reportFailure(stderr, nil, "configuration")
		return processcode.InvalidStartup
	}
	if err := settings.MemoryLimit().ValidateRuntime(debug.SetMemoryLimit(-1)); err != nil {
		reportFailure(stderr, nil, "configuration")
		return processcode.InvalidStartup
	}
	info, err := buildinfo.ReadWorker(settings.LongRunning().Runtime().Environment())
	if err != nil {
		reportFailure(stderr, nil, "build_contract")
		return processcode.InvalidStartup
	}
	embedded, err := loadContract()
	if err != nil {
		reportFailure(stderr, nil, "build_contract")
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
	return startWorker(process, settings, info, embedded, logger, telemetryRuntime)
}

func loadSchemaContract() (schemacontract.EmbeddedContract, error) {
	registry, err := migration.LoadReleaseRegistry()
	if err != nil {
		return schemacontract.EmbeddedContract{}, err
	}
	return schemacontract.LoadEmbedded(registry)
}
