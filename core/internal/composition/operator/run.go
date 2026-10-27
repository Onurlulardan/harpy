package operator

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
	return run(process.StartupContext(), arguments, os.Stdout, os.Stderr).Int()
}

func run(
	ctx context.Context,
	arguments []string,
	stdout io.Writer,
	stderr io.Writer,
) processcode.Code {
	arguments, err := envfile.Consume(arguments, "operator_initialize")
	if err != nil {
		writeFailureCause(stderr, "configuration", processcode.InvalidStartup, err.Error())
		return processcode.InvalidStartup
	}
	if len(arguments) == 1 && arguments[0] == "local-artifacts" {
		return runLocalAction(stderr, "local_artifacts", runLocalArtifacts())
	}
	if len(arguments) == 1 && arguments[0] == "local-license" {
		return runLocalAction(stderr, "local_license", runLocalLicense(ctx))
	}
	if len(arguments) == 1 && arguments[0] == "publish-local-authority" {
		return runLocalAction(
			stderr, "publish_local_authority", runLocalAuthorityPublication(ctx),
		)
	}
	if len(arguments) == 2 && arguments[0] == "local-delivery" {
		return runLocalAction(stderr, "local_delivery",
			runLocalDelivery(ctx, arguments[1], stdout))
	}
	if len(arguments) == 1 && arguments[0] == "reconcile-restored-identity" {
		return runRestoredIdentityCommand(ctx, stdout, stderr)
	}
	if len(arguments) == 1 && arguments[0] == "reconcile-restored-license" {
		return runRestoredLicenseCommand(ctx, stdout, stderr)
	}
	if len(arguments) == 1 && arguments[0] == "await-restored-reactivation" {
		return runPostRestoreObserver(ctx, stdout, stderr)
	}
	if len(arguments) != 1 || arguments[0] != "initialize" {
		writeFailure(stderr, "invocation", processcode.InvalidStartup)
		return processcode.InvalidStartup
	}
	result, stage, err := initializeEnvironment(ctx)
	if err != nil {
		code := classifyStartupError(stage, err)
		writeFailureCause(stderr, stage, code, err.Error())
		return code
	}
	if !result.IsValid() || writeSuccess(stdout, result) != nil {
		writeFailure(stderr, "result", processcode.OperatorTerminal)
		return processcode.OperatorTerminal
	}
	return processcode.Clean
}

func runLocalAction(
	stderr io.Writer,
	stage string,
	err error,
) processcode.Code {
	if err == nil {
		return processcode.Clean
	}
	writeFailure(stderr, stage, processcode.OperatorResumable)
	return processcode.OperatorResumable
}

func classifyStartupError(stage string, err error) processcode.Code {
	switch stage {
	case "configuration", "build_contract", "manifest":
		return processcode.InvalidStartup
	case "mode", "database", "composition", "authority":
		return processcode.DependencyPreflight
	case "recovery_signing":
		return classifyRecoveryError(err)
	default:
		return classifyInitializeError(err)
	}
}
