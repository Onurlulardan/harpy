package migrator

import (
	"context"
	"fmt"
	"io"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/baseline"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/migratoroutput"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/randomid"
)

func runBaseline(
	ctx context.Context,
	action deployAction,
	stdout io.Writer,
	stderr io.Writer,
) processcode.Code {
	switch action {
	case deployUp:
		return runBaselineUp(ctx, stdout, stderr)
	case deployPlan:
		return runBaselinePlan(ctx, stdout, stderr)
	case deployVerify:
		return runBaselineVerify(ctx, stdout, stderr)
	default:
		reportFailure(stderr, nil, "invocation", string(action))
		return processcode.InvalidStartup
	}
}

func runBaselinePlan(
	ctx context.Context,
	stdout io.Writer,
	stderr io.Writer,
) processcode.Code {
	settings, err := config.LoadMigratorDeploy()
	if err != nil {
		reportFailureCause(stderr, nil, "configuration", "plan", err)
		return processcode.InvalidStartup
	}
	appliedVersions, err := baseline.AppliedPlatformVersions(
		ctx, settings.Runtime().Postgres().DSN().Value(),
	)
	if err != nil {
		reportFailureCause(stderr, nil, "migration", "plan", err)
		return processcode.DeployMigration
	}
	appliedSet := make(map[int64]struct{}, len(appliedVersions))
	current := int64(0)
	for _, version := range appliedVersions {
		appliedSet[version] = struct{}{}
		if version > current { current = version }
	}
	pending := []runtimev1.MigrationPlanEntry{}
	for _, migration := range baseline.PlatformMigrations() {
		if _, exists := appliedSet[migration.Version]; exists { continue }
		pending = append(pending, runtimev1.MigrationPlanEntry{
			Owner: "platform_migrator", Scope: "platform", Version: migration.Version,
			TransactionMode: "transactional", Phase: "baseline",
			TargetDatabaseGeneration: 1, NextMinimumRuntimeGeneration: 1,
			Origin: "fresh", Path: migration.Path, SHA256: migration.SHA256,
		})
	}
	result, err := migratoroutput.NewBaselinePlan(current, pending)
	if err != nil {
		reportFailureCause(stderr, nil, "runtime", "plan", err)
		return processcode.ServeFailure
	}
	if err := result.Write(stdout); err != nil {
		reportFailureCause(stderr, nil, "runtime", "plan", err)
		return processcode.ServeFailure
	}
	return processcode.Clean
}

func runBaselineVerify(
	ctx context.Context,
	stdout io.Writer,
	stderr io.Writer,
) processcode.Code {
	settings, err := config.LoadMigratorDeploy()
	if err != nil {
		reportFailureCause(stderr, nil, "configuration", "verify", err)
		return processcode.InvalidStartup
	}
	appliedVersions, err := baseline.AppliedPlatformVersions(
		ctx, settings.Runtime().Postgres().DSN().Value(),
	)
	if err != nil || len(appliedVersions) != len(baseline.PlatformMigrations()) {
		if err == nil { err = fmt.Errorf("platform baseline is incomplete") }
		reportFailureCause(stderr, nil, "migration", "verify", err)
		return processcode.DeployMigration
	}
	runID, err := randomid.NewUUIDv4()
	if err != nil {
		reportFailureCause(stderr, nil, "runtime", "verify", err)
		return processcode.ServeFailure
	}
	result, err := migratoroutput.NewBaselineVerify(runID, baseline.LatestPlatformVersion())
	if err != nil {
		reportFailureCause(stderr, nil, "runtime", "verify", err)
		return processcode.ServeFailure
	}
	if err := result.Write(stdout); err != nil {
		reportFailureCause(stderr, nil, "runtime", "verify", err)
		return processcode.ServeFailure
	}
	return processcode.Clean
}

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
	result, err := migratoroutput.NewBaselineUp(
		runID, applied.AppliedVersions, applied.Version,
	)
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
