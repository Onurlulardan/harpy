package systemtransfer

import (
	"context"
	"errors"
)

func (worker *ExportWorker) finishOrRetry(
	ctx context.Context,
	claim ExportJobClaim,
	attempt ExportAttempt,
	cause error,
) (ExportExecutionResult, error) {
	cause = normalizeExportCommitError(cause)
	if errors.Is(cause, ErrExportLeaseLost) {
		return ExportExecutionResult{State: ExportExecutionLeaseLost}, cause
	}
	state, code, permanent := classifyExportFailure(cause)
	if !permanent && claim.AttemptNumber < claim.MaximumAttempts {
		return ExportExecutionResult{State: ExportExecutionRetry,
			Code: ExportCodeDependencyUnavailable}, cause
	}
	if !permanent {
		state, code = ExportFinalFailed, ExportCodeDependencyUnavailable
	}
	now, err := worker.exportNow()
	if err != nil {
		return ExportExecutionResult{}, errors.Join(cause, err)
	}
	finalizeContext, cancel := exportFinalizationContext(ctx)
	defer cancel()
	finishErr := worker.scopes.InFreshExportCheckpoint(finalizeContext, attempt,
		func(scoped context.Context, scope ExportScope) error {
			if scope.TenantID() != attempt.TenantID || scope.ActorID() != attempt.ActorID {
				return ErrExportStateConflict
			}
			store, bindErr := worker.stores.BindExportWorker(scoped, scope.IAMScope())
			if bindErr != nil || nilDependency(store) {
				return errors.Join(ErrExportUnavailable, bindErr)
			}
			return store.PersistOutboundFailure(scoped, ExportFailure{Attempt: attempt,
				State: state, Code: code, At: now})
		})
	if finishErr != nil {
		return ExportExecutionResult{}, errors.Join(cause, finishErr)
	}
	if state == ExportFinalCancelled {
		return ExportExecutionResult{State: ExportExecutionCancelled, Code: code}, nil
	}
	return ExportExecutionResult{State: ExportExecutionFailed, Code: code}, nil
}

func classifyExportFailure(err error) (ExportFinalState, string, bool) {
	cases := []struct {
		sentinel error
		state    ExportFinalState
		code     string
	}{{ErrExportCancelled, ExportFinalCancelled, ExportCodeCancelled},
		{ErrExportAuthorizationRevoked, ExportFinalFailed, ExportCodeAuthorizationRevoked},
		{ErrExportLimit, ExportFinalFailed, ExportCodeLimitExceeded},
		{ErrExportQueryInvalid, ExportFinalFailed, ExportCodeQueryInvalid},
		{ErrExportCatalogPinStale, ExportFinalFailed, ExportCodeCatalogPinStale},
		{ErrExportArtifactTampered, ExportFinalFailed, ExportCodeArtifactTampered},
		{ErrExportStateConflict, ExportFinalFailed, ExportCodeStateConflict}}
	for _, candidate := range cases {
		if errors.Is(err, candidate.sentinel) {
			return candidate.state, candidate.code, true
		}
	}
	return ExportFinalFailed, ExportCodeDependencyUnavailable, false
}
