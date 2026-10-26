package migrator

import (
	"context"
	"errors"
	"fmt"
)

func (runner *Runner) Up(
	ctx context.Context,
	request UpRequest,
) (_ UpResult, resultErr error) {
	if ctx == nil {
		return UpResult{}, fmt.Errorf("up context is nil")
	}
	provider, err := runner.buildGooseProvider(request.RunID)
	if err != nil {
		return UpResult{}, err
	}
	lock, err := runner.acquirePlatformLock(ctx)
	if err != nil {
		return UpResult{}, err
	}
	defer func() {
		resultErr = errors.Join(resultErr, runner.releasePlatformLock(ctx, lock))
	}()
	initial, _, err := readHistory(ctx, lock.connection)
	if err != nil {
		return UpResult{}, err
	}
	initialVersion, initialPending, err := classifyHistory(
		initial, runner.selected, runner.historyAliases,
	)
	if err != nil {
		return UpResult{}, err
	}
	if recovered, ok, err := runner.recoverFinalizedRelease(
		ctx, lock.connection, initialVersion, len(initialPending),
	); err != nil {
		return UpResult{}, err
	} else if ok {
		return recovered, nil
	}
	if err := runner.bootstrapRunLifecycle(ctx, lock.connection, provider); err != nil {
		return UpResult{}, runner.handleBootstrapFailure(
			ctx, lock.connection, request.RunID, err,
		)
	}
	run, err := runner.ensureRunRunning(ctx, lock.connection, request.RunID)
	if err != nil {
		run, err = runner.reconcileRunStart(ctx, lock.connection, request.RunID, err)
		if err != nil {
			return UpResult{}, err
		}
	}
	if run.state == "succeeded" {
		return runner.resultForSucceededRun(ctx, lock.connection, request.RunID, initial)
	}
	if ctx.Err() != nil {
		return UpResult{}, runner.finalizeFailure(
			ctx, lock.connection, request.RunID, "migration_cancelled", ctx.Err(),
		)
	}
	_, migrationErr := provider.Up(ctx)
	reconcileContext, cancelReconcile := runner.reconcileContext(ctx)
	defer cancelReconcile()
	after, _, historyErr := readHistory(reconcileContext, lock.connection)
	if historyErr != nil {
		return UpResult{}, ambiguousOutcome(
			request.RunID, "reconcile migrations", errors.Join(migrationErr, historyErr),
		)
	}
	current, pending, verifyErr := classifyHistory(
		after, runner.selected, runner.historyAliases,
	)
	if verifyErr != nil {
		return UpResult{}, runner.finalizeFailure(
			ctx, lock.connection, request.RunID,
			"migration_verification_failed", verifyErr,
		)
	}
	result := UpResult{
		RunID: request.RunID, DatabaseVersion: current,
		AppliedVersions: appliedVersionDifference(initial, after, runner.selected),
	}
	if migrationErr != nil && len(pending) != 0 {
		code := "migration_execution_failed"
		if ctx.Err() != nil {
			code = "migration_cancelled"
		}
		return UpResult{}, runner.finalizeFailure(
			ctx, lock.connection, request.RunID, code, migrationErr,
		)
	}
	if len(pending) != 0 {
		failure := fmt.Errorf("Goose returned with %d pending migrations", len(pending))
		return UpResult{}, runner.finalizeFailure(
			ctx, lock.connection, request.RunID, "migration_verification_failed", failure,
		)
	}
	tenantApplied, tenantErr := runner.runTenantMigrations(
		ctx, lock.connection, request.RunID,
	)
	if tenantErr != nil && !errors.Is(tenantErr, ErrCommitted) {
		if errors.Is(tenantErr, ErrOutcomeAmbiguous) {
			return UpResult{}, tenantErr
		}
		return UpResult{}, runner.finalizeFailure(
			ctx, lock.connection, request.RunID,
			"tenant_migration_failed", tenantErr,
		)
	}
	result.AppliedVersions = mergeAppliedVersions(result.AppliedVersions, tenantApplied)
	if err := runner.finalizeSuccess(ctx, lock.connection, request.RunID); err != nil {
		if isReleaseFinalizationError(err) {
			return UpResult{}, runner.finalizeFailure(
				ctx, lock.connection, request.RunID,
				"release_finalization_failed", err,
			)
		}
		if reconcileErr := runner.reconcileSuccess(
			ctx, lock.connection, request.RunID, err,
		); reconcileErr != nil {
			return UpResult{}, reconcileErr
		}
	}
	if migrationErr != nil || tenantErr != nil {
		return result, &OutcomeError{
			RunID: request.RunID, Stage: "up",
			Cause: errors.Join(ErrCommitted, migrationErr, tenantErr),
		}
	}
	if ctx.Err() != nil {
		return result, &OutcomeError{
			RunID: request.RunID, Stage: "up",
			Cause: errors.Join(ErrCommitted, ctx.Err()),
		}
	}
	return result, nil
}
