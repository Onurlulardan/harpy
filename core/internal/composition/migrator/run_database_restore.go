package migrator

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

func runDatabaseRestore(
	ctx context.Context,
	stderr io.Writer,
) processcode.Code {
	settings, err := config.LoadDatabaseRestore()
	if err != nil {
		reportFailure(stderr, nil, "configuration", "database-restore")
		return processcode.InvalidStartup
	}
	runtime := settings.Runtime()
	info, err := buildinfo.ReadMigrator(runtime.Environment())
	if err != nil {
		reportFailure(stderr, nil, "build_contract", "database-restore")
		return processcode.InvalidStartup
	}
	embedded, err := loadSchemaContract()
	if err != nil {
		reportFailure(stderr, nil, "build_contract", "database-restore")
		return processcode.InvalidStartup
	}
	logger, err := telemetry.NewLogger(stderr, runtime.LogLevel(), info)
	if err != nil {
		return processcode.InvalidStartup
	}
	telemetryRuntime, err := telemetry.New(
		ctx, runtime.Telemetry(), info, logger,
	)
	if err != nil {
		reportFailure(stderr, logger, "configuration", "database-restore")
		return processcode.InvalidStartup
	}
	defer func() {
		shutdown, cancel := context.WithTimeout(
			context.Background(), runtime.Telemetry().ExportTimeout(),
		)
		defer cancel()
		_ = telemetryRuntime.Shutdown(shutdown)
	}()
	pool, err := postgres.Open(
		ctx, runtime.ControlPostgres(), postgres.ProfileDatabaseRestore, embedded,
	)
	if err != nil {
		reportFailure(stderr, logger, "dependency_preflight", "database-restore")
		return processcode.DependencyPreflight
	}
	defer pool.Close()
	readiness := pool.ProbeReadiness(ctx)
	if !readiness.DatabaseReady() || !readiness.DatabaseRoleReady() ||
		!readiness.SchemaContractReady() {
		reportFailure(stderr, logger, "dependency_preflight", "database-restore")
		return processcode.DependencyPreflight
	}
	err = platformops.RunDatabaseRestore(
		ctx, pool, settings, embedded,
		platformops.RecoveryFactories{
			Documents:         documents.NewRecovery,
			Targets:           tenancy.NewRecoveryTargets,
			IAMEvidence:       iam.NewRecoveryEvidence,
			AuthorityEvidence: platformauthority.NewRecoveryEvidence,
		},
		func(work context.Context) (any, error) {
			connect, cancel := context.WithTimeout(
				work, runtime.ScratchPostgres().ConnectTimeout()+5*time.Second,
			)
			defer cancel()
			scratch, openErr := postgres.Open(
				connect, runtime.ScratchPostgres(),
				postgres.ProfileDatabaseRestore, embedded,
			)
			if openErr != nil {
				return nil, openErr
			}
			ready := scratch.ProbeReadiness(connect)
			if !ready.DatabaseReady() || !ready.DatabaseRoleReady() ||
				!ready.SchemaContractReady() {
				scratch.Close()
				return nil, fmt.Errorf("restored schema contract differs")
			}
			return scratch, nil
		},
	)
	if err == nil {
		return processcode.Clean
	}
	code, class := classifyDatabaseRestoreFailure(err)
	reportFailure(stderr, logger, class, "database-restore")
	return code
}

func classifyDatabaseRestoreFailure(
	err error,
) (processcode.Code, string) {
	switch platformops.ClassifyDatabaseRestoreError(err) {
	case platformops.RestoreFailureInvalid:
		return processcode.InvalidStartup, "configuration"
	case platformops.RestoreFailureDependency:
		return processcode.DependencyPreflight, "dependency_preflight"
	case platformops.RestoreFailureTerminal:
		return processcode.OperatorTerminal, "restore_terminal"
	default:
		return processcode.OperatorResumable, "restore_resumable"
	}
}
