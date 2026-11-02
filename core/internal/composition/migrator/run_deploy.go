package migrator

import (
	"context"
	"errors"
	"io"

	platformmigrator "github.com/Onurlulardan/harpy/core/internal/platform/migrator"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
)

func runDeploy(
	ctx context.Context,
	action deployAction,
	stdout io.Writer,
	stderr io.Writer,
) processcode.Code {
	return runBaseline(ctx, action, stdout, stderr)
}

func classifyDeployFailure(err error) (processcode.Code, string) {
	if errors.Is(err, errDeployRuntime) {
		return processcode.ServeFailure, "runtime"
	}
	if errors.Is(err, platformmigrator.ErrDependencyPreflight) {
		return processcode.DependencyPreflight, "dependency_preflight"
	}
	return processcode.DeployMigration, "migration"
}
