package migrator

import (
	"context"
	"io"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/cutover"
	cutoverpostgres "github.com/Onurlulardan/harpy/core/internal/platform/cutover/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
)

func runCutoverFinalization(
	ctx context.Context,
	stdout io.Writer,
	stderr io.Writer,
) processcode.Code {
	settings, err := config.LoadCutoverFinalization()
	if err != nil {
		reportFailure(stderr, nil, "configuration", "cutover_finalization")
		return processcode.InvalidStartup
	}
	runtime, err := startCutoverRuntime(
		ctx, settings.Environment(), settings.LogLevel(), settings.Telemetry(), stderr,
	)
	if err != nil {
		reportFailure(stderr, nil, "runtime_contract", "cutover_finalization")
		return processcode.InvalidStartup
	}
	defer func() { _ = runtime.shutdown() }()
	evidence, err := cutover.LoadFinalizationEvidence(settings.Artifacts(), time.Now().UTC())
	if err != nil {
		reportFailure(stderr, runtime.logger, "dependency_preflight", "cutover_finalization")
		return processcode.DependencyPreflight
	}
	leaseCtx, cancel := context.WithDeadline(ctx, evidence.FinalizerLease.ExpiresAt())
	defer cancel()
	database, err := cutoverpostgres.OpenFinalizationDatabase(
		leaseCtx, settings.Postgres(), evidence.FinalizerLease,
	)
	if err != nil {
		reportFailure(stderr, runtime.logger, "dependency_preflight", "cutover_finalization")
		return processcode.DependencyPreflight
	}
	defer func() { _ = database.Close() }()
	result, err := database.Finalize(leaseCtx, cutoverpostgres.FinalizationInput{
		AdoptionLease: evidence.AdoptionLease, SessionSeal: evidence.SessionSeal,
		Cleanup: evidence.Cleanup, Authority: evidence.Authority,
		FinalizerLease: evidence.FinalizerLease, TerminalEvidence: evidence.TerminalEvidence,
	})
	if err != nil {
		reportFailure(stderr, runtime.logger, "migration", "cutover_finalization")
		return processcode.DeployMigration
	}
	if database.Close() != nil || runtime.shutdown() != nil {
		reportFailure(stderr, runtime.logger, "runtime", "cutover_finalization")
		return processcode.ServeFailure
	}
	code, err := writeFinalizationResult(
		stdout, evidence.FinalizerLease.OperationID(), evidence.FinalizerLease.RunID(),
		result.TerminalState, result.TerminalEvidenceSHA256,
	)
	if err != nil {
		reportFailure(stderr, runtime.logger, "runtime", "cutover_finalization")
		return processcode.ServeFailure
	}
	return code
}
