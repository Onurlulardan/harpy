package migrator

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/orchestration/v1adoption"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/cutover"
	cutoverpostgres "github.com/Onurlulardan/harpy/core/internal/platform/cutover/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
)

func runCutoverAdoption(
	ctx context.Context,
	stdout io.Writer,
	stderr io.Writer,
) processcode.Code {
	settings, err := config.LoadCutoverAdoption()
	if err != nil {
		reportFailure(stderr, nil, "configuration", "cutover_adoption")
		return processcode.InvalidStartup
	}
	runtime, err := startCutoverRuntime(
		ctx, settings.Environment(), settings.LogLevel(), settings.Telemetry(), stderr,
	)
	if err != nil {
		reportFailure(stderr, nil, "runtime_contract", "cutover_adoption")
		return processcode.InvalidStartup
	}
	defer func() { _ = runtime.shutdown() }()
	evidence, err := cutover.LoadAdoptionEvidence(settings.Artifacts(), time.Now().UTC())
	if err != nil {
		reportFailure(stderr, runtime.logger, "dependency_preflight", "cutover_adoption")
		return processcode.DependencyPreflight
	}
	leaseCtx, cancel := context.WithDeadline(ctx, evidence.Lease.ExpiresAt())
	defer cancel()
	applyRegistry, err := newCutoverApplyRegistry()
	if err != nil {
		reportFailure(stderr, runtime.logger, "runtime_contract", "cutover_adoption")
		return processcode.InvalidStartup
	}
	payload := evidence.Batch.Payload()
	relation, found := evidence.Lease.Relation(payload.RelationOrdinal)
	if !found || v1adoption.ValidateCutoverBatch(relation.ProcedureContractID, payload) != nil {
		reportFailure(stderr, runtime.logger, "dependency_preflight", "cutover_adoption")
		return processcode.DependencyPreflight
	}
	database, err := cutoverpostgres.OpenAdoptionDatabase(
		leaseCtx, settings.SourcePostgres(), settings.TargetPostgres(), evidence.Lease,
		applyRegistry,
	)
	if err != nil {
		reportFailure(stderr, runtime.logger, "dependency_preflight", "cutover_adoption")
		return processcode.DependencyPreflight
	}
	defer func() { _ = database.Close() }()
	runtime.logger.Info(
		"cutover sessions opened",
		slog.String("operation_id", evidence.Lease.OperationID()),
		slog.String("run_id", evidence.Lease.RunID()),
		slog.Int("source_sessions", evidence.Lease.SourceConnectionLimit()),
		slog.Int("target_sessions", evidence.Lease.TargetConnectionLimit()),
	)
	seal, err := cutover.AwaitSessionSeal(
		leaseCtx, settings.Artifacts().SessionSealReceiptPath().Value(),
		evidence, func() time.Time { return time.Now().UTC() },
	)
	if err != nil || database.VerifySessionSeal(leaseCtx, evidence.Lease, seal) != nil {
		reportFailure(stderr, runtime.logger, "dependency_preflight", "cutover_adoption")
		return processcode.DependencyPreflight
	}
	result, err := database.Apply(leaseCtx, evidence.Lease, evidence.Batch)
	if err != nil {
		reportFailure(stderr, runtime.logger, "migration", "cutover_adoption")
		return processcode.DeployMigration
	}
	if database.Close() != nil || runtime.shutdown() != nil {
		reportFailure(stderr, runtime.logger, "runtime", "cutover_adoption")
		return processcode.ServeFailure
	}
	if err := writeAdoptionResult(
		stdout, evidence.Lease.OperationID(), evidence.Lease.RunID(),
		evidence.Batch.SHA256(), result.AppliedCount,
		result.KeyStateSHA256, result.StateSHA256,
	); err != nil {
		reportFailure(stderr, runtime.logger, "runtime", "cutover_adoption")
		return processcode.ServeFailure
	}
	return processcode.Clean
}
