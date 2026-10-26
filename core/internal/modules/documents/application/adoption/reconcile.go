package adoption

import "context"

func (application *Application) Reconcile(ctx context.Context,
	request ReconcileRequest,
) (Run, error) {
	if application == nil || ctx == nil || ctx.Err() != nil ||
		!validAuthority(request.Authority) || !validLSN(request.FinalLSN) ||
		!validLSN(request.AppliedLSN) ||
		request.OccurredAt.IsZero() {
		return Run{}, ErrInvalid
	}
	result, err := application.store.Reconcile(ctx, request)
	if err != nil {
		return Run{}, mapError(err)
	}
	if !validRun(result, request.Authority) {
		return Run{}, ErrUnavailable
	}
	if result.State != "quiesced" || result.SourceCount != result.TargetCount ||
		result.SourceBytes != result.TargetBytes || result.SourceMerkleSHA256 == [32]byte{} ||
		result.SourceMerkleSHA256 != result.TargetMerkleSHA256 ||
		result.PendingObjectCount != 0 || result.PendingJobCount != 0 ||
		result.FinalLSN != result.AppliedLSN {
		return Run{}, ErrBlocked
	}
	return result, nil
}
