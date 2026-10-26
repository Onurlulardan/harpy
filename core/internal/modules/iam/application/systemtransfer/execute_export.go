package systemtransfer

import (
	"context"
	"errors"
)

func (worker *ExportWorker) ExecuteExport(
	ctx context.Context,
	claim ExportJobClaim,
) (ExportExecutionResult, error) {
	if worker == nil || ctx == nil || claim.Validate() != nil {
		return ExportExecutionResult{}, ErrExportWorkerInvalid
	}
	loadContext, cancel := exportLoadContext(ctx)
	defer cancel()
	attempt, err := worker.loader.LoadExportAttempt(loadContext, claim)
	if err != nil {
		return worker.unloadedFailure(claim, err)
	}
	if attempt.CheckOwner() != nil || matchExportClaim(attempt, claim) != nil {
		return ExportExecutionResult{State: ExportExecutionFailed,
			Code: ExportCodeStateConflict}, ErrExportStateConflict
	}
	switch attempt.State {
	case ExportReady:
		return ExportExecutionResult{State: ExportExecutionReady}, nil
	case ExportCancelled:
		return ExportExecutionResult{State: ExportExecutionCancelled,
			Code: ExportCodeCancelled}, nil
	case ExportFailed, ExportExpired:
		return ExportExecutionResult{State: ExportExecutionFailed,
			Code: ExportCodeStateConflict}, nil
	case ExportOutcomeUnknown:
		return ExportExecutionResult{State: ExportExecutionOutcomeUnknown,
			Code: ExportCodeOutcomeUnknown}, nil
	}
	attempt.AttemptID = claim.AttemptID
	attempt.LeaseOwnerID = claim.LeaseOwnerID
	attempt.AttemptGeneration = claim.LeaseGeneration
	if attempt.Validate() != nil {
		return ExportExecutionResult{State: ExportExecutionFailed,
			Code: ExportCodeStateConflict}, ErrExportStateConflict
	}
	if err = worker.checkLease(ctx, claim); err != nil {
		if errors.Is(err, ErrExportCancelled) {
			return worker.finishOrRetry(ctx, claim, attempt, err)
		}
		return leaseExecutionResult(err)
	}
	claimed, err := worker.claimExport(ctx, attempt)
	if err != nil {
		return worker.finishOrRetry(ctx, claim, attempt, err)
	}
	if claimed.State == ExportReady {
		return ExportExecutionResult{State: ExportExecutionReady}, nil
	}
	if err = worker.streamExport(ctx, claim, attempt); err != nil {
		if errors.Is(err, ErrExportOutcomeUnknown) {
			result := ExportExecutionResult{State: ExportExecutionOutcomeUnknown,
				Code: ExportCodeOutcomeUnknown}
			if errors.Is(err, ErrExportOutcomeUnknownPersist) {
				return result, err
			}
			return result, nil
		}
		return worker.finishOrRetry(ctx, claim, attempt, err)
	}
	return ExportExecutionResult{State: ExportExecutionReady}, nil
}

func (worker *ExportWorker) unloadedFailure(
	claim ExportJobClaim,
	cause error,
) (ExportExecutionResult, error) {
	if errors.Is(cause, ErrExportLeaseLost) || errors.Is(cause, ErrExportCancelled) {
		return leaseExecutionResult(cause)
	}
	if claim.AttemptNumber < claim.MaximumAttempts {
		return ExportExecutionResult{State: ExportExecutionRetry,
			Code: ExportCodeDependencyUnavailable}, cause
	}
	return ExportExecutionResult{State: ExportExecutionFailed,
		Code: ExportCodeDependencyUnavailable}, cause
}

func leaseExecutionResult(err error) (ExportExecutionResult, error) {
	if errors.Is(err, ErrExportCancelled) {
		return ExportExecutionResult{State: ExportExecutionCancelled,
			Code: ExportCodeCancelled}, err
	}
	return ExportExecutionResult{State: ExportExecutionLeaseLost}, err
}
