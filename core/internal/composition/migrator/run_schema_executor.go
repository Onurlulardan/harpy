package migrator

import (
	"context"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

type schemaContractLoader func() (schemacontract.EmbeddedContract, error)

func runSchemaExecutor(ctx context.Context, stderr io.Writer) processcode.Code {
	return runSchemaExecutorProcessWithContractLoader(
		lifecycle.LocalProcess(ctx), stderr, loadSchemaContract,
	)
}

func runSchemaExecutorWithContractLoader(
	ctx context.Context,
	stderr io.Writer,
	loadContract schemaContractLoader,
) processcode.Code {
	return runSchemaExecutorProcessWithContractLoader(
		lifecycle.LocalProcess(ctx), stderr, loadContract,
	)
}

func runSchemaExecutorProcess(
	process *lifecycle.Process,
	stderr io.Writer,
) processcode.Code {
	return runSchemaExecutorProcessWithContractLoader(
		process, stderr, loadSchemaContract,
	)
}

func runSchemaExecutorProcessWithContractLoader(
	process *lifecycle.Process,
	stderr io.Writer,
	loadContract schemaContractLoader,
) processcode.Code {
	ctx := process.StartupContext()
	settings, err := config.LoadSchemaExecutor()
	if err != nil {
		reportFailure(stderr, nil, "configuration", "schema-executor")
		return processcode.InvalidStartup
	}
	runtimeSettings := settings.LongRunning().Runtime()
	info, err := buildinfo.ReadMigrator(runtimeSettings.Environment())
	if err != nil {
		reportFailure(stderr, nil, "build_contract", "schema-executor")
		return processcode.InvalidStartup
	}
	embedded, err := loadContract()
	if err != nil {
		reportFailure(stderr, nil, "build_contract", "schema-executor")
		return processcode.InvalidStartup
	}
	logger, err := telemetry.NewLogger(stderr, runtimeSettings.LogLevel(), info)
	if err != nil {
		reportFailure(stderr, nil, "configuration", "schema-executor")
		return processcode.InvalidStartup
	}
	telemetryRuntime, err := telemetry.New(ctx, runtimeSettings.Telemetry(), info, logger)
	if err != nil {
		reportFailure(nil, logger, "configuration", "schema-executor")
		return processcode.InvalidStartup
	}
	return startSchemaExecutor(
		process, settings, info, embedded, logger, telemetryRuntime,
	)
}

func loadSchemaContract() (schemacontract.EmbeddedContract, error) {
	registry, err := migration.LoadReleaseRegistry()
	if err != nil {
		return schemacontract.EmbeddedContract{}, err
	}
	return schemacontract.LoadEmbedded(registry)
}
