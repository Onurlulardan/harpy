package recovery

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
	_ = stdout
	preview, err := parseInvocation(commandArguments(arguments))
	if err != nil {
		reportFailureCause(stderr, nil, "invocation", "invalid", err)
		return processcode.InvalidStartup
	}
	profile := "database_backup_executor"
	if preview.mode == invocationDatabaseRestore {
		profile = "database_restore"
	}
	arguments, err = envfile.Consume(arguments, profile)
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
	case invocationDatabaseBackupExecutor:
		return runDatabaseBackupExecutorProcess(process, stderr)
	case invocationDatabaseRestore:
		return runDatabaseRestore(process.StartupContext(), stderr)
	default:
		reportFailureCause(stderr, nil, "invocation", "invalid", invalidInvocation())
		return processcode.InvalidStartup
	}
}
