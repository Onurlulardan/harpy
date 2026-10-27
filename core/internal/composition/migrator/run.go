package migrator

import (
	"context"
	"io"
	"os"

	"github.com/Onurlulardan/harpy/core/internal/platform/envfile"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
)

func Run(arguments []string) int {
	process := lifecycle.NotifyProcess(context.Background())
	defer process.Stop()
	return runProcess(process, arguments, os.Stdout, os.Stderr).Int()
}

func run(
	ctx context.Context,
	arguments []string,
	stdout io.Writer,
	stderr io.Writer,
) processcode.Code {
	return runProcess(lifecycle.LocalProcess(ctx), arguments, stdout, stderr)
}

func runProcess(
	process *lifecycle.Process,
	arguments []string,
	stdout io.Writer,
	stderr io.Writer,
) processcode.Code {
	arguments, err := envfile.Consume(arguments, "migrator_deploy")
	if err != nil {
		reportFailureCause(stderr, nil, "configuration", "invalid", err)
		return processcode.InvalidStartup
	}
	invocation, err := parseInvocation(arguments)
	if err != nil {
		reportFailureCause(stderr, nil, "invocation", "invalid", err)
		return processcode.InvalidStartup
	}
	switch invocation.mode {
	case invocationSchemaExecutor:
		return runSchemaExecutorProcess(process, stderr)
	case invocationTenantMigrationExecutor:
		return runTenantMigrationExecutorProcess(process, stderr)
	case invocationDatabaseBackupExecutor:
		return runDatabaseBackupExecutorProcess(process, stderr)
	case invocationDatabaseRestore:
		return runDatabaseRestore(process.StartupContext(), stderr)
	case invocationCutoverAdoption:
		return runCutoverAdoption(process.StartupContext(), stdout, stderr)
	case invocationCutoverFinalization:
		return runCutoverFinalization(process.StartupContext(), stdout, stderr)
	default:
		return runDeploy(process.StartupContext(), invocation.action, stdout, stderr)
	}
}
