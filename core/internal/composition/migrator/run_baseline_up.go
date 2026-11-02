package migrator

import (
	"context"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/platform/baseline"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/migratoroutput"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/randomid"
)

func runBaselineUp(
	ctx context.Context,
	stdout io.Writer,
	stderr io.Writer,
) processcode.Code {
	settings, err := config.LoadMigratorDeploy()
	if err != nil {
		reportFailureCause(stderr, nil, "configuration", "up", err)
		return processcode.InvalidStartup
	}
	applied, err := baseline.ApplyPlatform(ctx, settings.Runtime().Postgres().DSN().Value())
	if err != nil {
		reportFailureCause(stderr, nil, "migration", "up", err)
		return processcode.DeployMigration
	}
	runID, err := randomid.NewUUIDv4()
	if err != nil {
		reportFailureCause(stderr, nil, "runtime", "up", err)
		return processcode.ServeFailure
	}
	versions := []int64{}
	if applied.Applied {
		versions = []int64{applied.Version}
	}
	result, err := migratoroutput.NewBaselineUp(runID, versions, applied.Version)
	if err != nil {
		reportFailureCause(stderr, nil, "runtime", "up", err)
		return processcode.ServeFailure
	}
	if err := result.Write(stdout); err != nil {
		reportFailureCause(stderr, nil, "runtime", "up", err)
		return processcode.ServeFailure
	}
	return processcode.Clean
}
