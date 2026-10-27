package recoveryorchestration

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func RunRestore(
	ctx context.Context,
	dependencies RestoreDependencies,
) error {
	executor, err := newRestoreExecutor(ctx, dependencies)
	if err != nil {
		return invalidRestore(err)
	}
	defer executor.close()
	now := time.Now().UTC().Truncate(time.Microsecond)
	attemptID, attemptErr := dependencies.NewAttemptID(now)
	leaseToken, leaseErr := dependencies.NewAttemptID(now)
	if attemptErr != nil || leaseErr != nil {
		return invalidRestore(fmt.Errorf("restore attempt identity is unavailable"))
	}
	identity := ports.RestoreExecutionIdentity{
		OperationID: executor.authority.planArtifact.OperationID(),
		AttemptID:   attemptID, LeaseToken: leaseToken,
	}
	claim, found, err := executor.execution.Claim(
		ctx, identity, executor.authority.planArtifact.SHA256(),
	)
	if err != nil || !found {
		return resumableRestore(fmt.Errorf("restore execution claim is unavailable"))
	}
	err = executor.withRestoreLease(ctx, identity, func(work context.Context) error {
		return executor.execute(work, claim, identity)
	})
	if err != nil && ctx.Err() == nil {
		class := restoreErrorClass(err)
		_ = executor.execution.RejectFailure(
			context.WithoutCancel(ctx), ports.RestoreFailure{
				RestoreExecutionIdentity: identity,
				Code:                     "restore_execution_failed",
				Fenced:                   class == restoreTerminal,
			},
		)
	}
	return err
}

func newRestoreExecutor(
	ctx context.Context,
	dependencies RestoreDependencies,
) (*databaseRestoreExecutor, error) {
	if dependencies.Execution == nil || dependencies.Storage == nil ||
		dependencies.Archive == nil || dependencies.OpenScratch == nil ||
		dependencies.NewAttemptID == nil || dependencies.NewDatabaseID == nil || dependencies.LocatorSHA256 == nil ||
		dependencies.Environment == "" || dependencies.ControlDSN == "" ||
		dependencies.ScratchDSN == "" {
		return nil, fmt.Errorf("database restore dependencies are unavailable")
	}
	trust, orchestrator, err := loadRestoreTrust(dependencies.Artifacts)
	if err != nil || trust.EnvironmentID() != dependencies.Environment {
		return nil, fmt.Errorf("restore trust identity differs")
	}
	authority, err := loadRestoreAuthority(
		ctx, dependencies.Artifacts, dependencies.Storage,
		trust, orchestrator, time.Now().UTC().Truncate(time.Microsecond),
	)
	if err != nil {
		return nil, err
	}
	keys, err := loadRestoreKeys(
		dependencies.Artifacts, dependencies.Storage, &authority,
	)
	if err != nil {
		authority.clear()
		return nil, err
	}
	if dependencies.Artifacts.Branch == RestoreRehydrate {
		if err = seedRehydrateControl(
			ctx, dependencies.Execution, authority,
		); err != nil {
			keys.clear()
			authority.clear()
			return nil, err
		}
	}
	return &databaseRestoreExecutor{
		execution: dependencies.Execution, storage: dependencies.Storage,
		authority: &authority, keys: keys, archive: dependencies.Archive,
		artifacts:     dependencies.Artifacts,
		controlDSN:    dependencies.ControlDSN,
		scratchDSN:    dependencies.ScratchDSN,
		openScratch:   dependencies.OpenScratch,
		newDatabaseID: dependencies.NewDatabaseID,
		locatorSHA256: dependencies.LocatorSHA256,
	}, nil
}

func (executor *databaseRestoreExecutor) close() {
	if executor == nil {
		return
	}
	executor.keys.clear()
	executor.authority.clear()
}
