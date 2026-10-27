package dashboardexport

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ExecuteClaimedRender is the only render execution surface. Its caller must
// reserve the exact encrypted spool before atomically claiming the attempt.
func (worker *Worker) ExecuteClaimedRender(
	ctx context.Context, attempt RenderDispatchAttempt, spool EncryptedSpool,
) (RenderExecutionResult, error) {
	if worker == nil || ctx == nil || spool == nil ||
		validateClaimedAttempt(attempt, spool) != nil {
		return RenderExecutionResult{}, ErrInvalid
	}
	if ctx.Err() != nil {
		return worker.finishPreEffectCancellation(ctx, attempt, spool)
	}
	claim := attempt.Claim
	if err := worker.authorize(ctx, claim.Pin, claim.ActorID); err != nil {
		return worker.finishRenderFailure(ctx, attempt, spool,
			"permanent", "AUTHORIZATION_REVOKED", false, uuid.Nil, [32]byte{})
	}
	pending, err := worker.artifacts.PrepareReserved(
		ctx, claim.Pin, claim.Generation, claim.LeaseOwner, spool, worker.renderer)
	if err != nil {
		if ctx.Err() != nil {
			return worker.finishPreEffectCancellation(ctx, attempt, spool)
		}
		return worker.finishRenderFailure(ctx, attempt, spool,
			"transient", "RENDER_FAILED", true, uuid.Nil, [32]byte{})
	}
	draft := pending.Draft()
	if err = worker.authorize(ctx, claim.Pin, claim.ActorID); err != nil {
		return worker.finishRenderFailure(ctx, attempt, spool,
			"permanent", "AUTHORIZATION_REVOKED", false, uuid.Nil, [32]byte{})
	}
	if ctx.Err() != nil {
		return worker.finishPreEffectCancellation(ctx, attempt, spool)
	}
	effectID, err := worker.ids.NewUUIDv7()
	if err != nil || !validUUIDv7(effectID) {
		return worker.finishRenderFailure(ctx, attempt, spool,
			"transient", "EFFECT_ID_UNAVAILABLE", true, uuid.Nil, [32]byte{})
	}
	requestSHA := renderUploadRequestSHA(draft)
	if err = worker.store.MarkEffectStarted(ctx, EffectStart{Attempt: renderAttemptIdentity(attempt),
		EffectID: effectID, EffectRequestSHA256: requestSHA, Draft: draft,
		StartedAt: worker.clock.Now()}); err != nil {
		if errors.Is(err, ErrDispatchLeaseStale) {
			return RenderExecutionResult{State: RenderExecutionLeaseLost}, err
		}
		return RenderExecutionResult{State: RenderExecutionOutcomeUnknown}, err
	}
	receipt, uploadErr := pending.Upload(ctx)
	if uploadErr != nil {
		if errors.Is(uploadErr, ErrOutcomeUnknown) || ctx.Err() != nil ||
			errors.Is(uploadErr, context.Canceled) ||
			errors.Is(uploadErr, context.DeadlineExceeded) {
			return worker.finishRenderUnknown(ctx, attempt, effectID, requestSHA)
		}
		return worker.finishRenderFailure(ctx, attempt, spool,
			"transient", "ARTIFACT_UPLOAD_FAILED", true, effectID, requestSHA)
	}
	return worker.finishRenderReady(ctx, attempt, spool, receipt, effectID, requestSHA)
}

func (worker *Worker) finishRenderReady(ctx context.Context, attempt RenderDispatchAttempt,
	spool EncryptedSpool, receipt UploadReceipt, effectID uuid.UUID, requestSHA [32]byte,
) (RenderExecutionResult, error) {
	receiptID, err := worker.ids.NewUUIDv7()
	if err != nil || !validUUIDv7(receiptID) || receipt.ObjectVersion == "" ||
		receipt.ProviderRequestID == "" || receipt.UploadedAt.IsZero() {
		return worker.finishRenderUnknown(ctx, attempt, effectID, requestSHA)
	}
	state, err := worker.store.CompleteReady(ctx, ReadyWrite{
		Attempt: renderAttemptIdentity(attempt), EffectID: effectID,
		EffectReceiptID: receiptID, EffectRequestSHA256: requestSHA,
		EffectResultSHA256: renderUploadResultSHA(receipt), Receipt: receipt,
		CompletedAt: worker.clock.Now()})
	if err != nil {
		return RenderExecutionResult{State: RenderExecutionOutcomeUnknown}, err
	}
	return finishKnownRender(spool, state)
}

func (worker *Worker) finishRenderFailure(ctx context.Context, attempt RenderDispatchAttempt,
	spool EncryptedSpool, class, code string, retryable bool,
	effectID uuid.UUID, requestSHA [32]byte,
) (RenderExecutionResult, error) {
	state, err := worker.store.FinishRender(ctx, RenderFailure{
		Attempt: renderAttemptIdentity(attempt), EffectID: effectID,
		EffectRequestSHA256: requestSHA, FailureClass: class, Code: code,
		FailureSHA256: failureSHA(code), Retryable: retryable,
		FinishedAt: worker.clock.Now()})
	if err != nil {
		return RenderExecutionResult{State: RenderExecutionOutcomeUnknown},
			fmt.Errorf("persist Analytics dashboard export failure: %w", err)
	}
	return finishKnownRender(spool, state)
}

func (worker *Worker) finishRenderUnknown(ctx context.Context, attempt RenderDispatchAttempt,
	effectID uuid.UUID, requestSHA [32]byte,
) (RenderExecutionResult, error) {
	err := worker.store.MarkOutcomeUnknown(ctx, RenderUnknown{
		Attempt: renderAttemptIdentity(attempt), EffectID: effectID,
		EffectRequestSHA256: requestSHA, ObservedAt: worker.clock.Now()})
	if err != nil {
		return RenderExecutionResult{State: RenderExecutionOutcomeUnknown},
			fmt.Errorf("persist Analytics dashboard export unknown outcome: %w", err)
	}
	return RenderExecutionResult{State: RenderExecutionOutcomeUnknown}, nil
}

func (worker *Worker) finishPreEffectCancellation(ctx context.Context,
	attempt RenderDispatchAttempt, spool EncryptedSpool,
) (RenderExecutionResult, error) {
	if errors.Is(context.Cause(ctx), ErrWorkerCancelled) {
		return worker.finishRenderFailure(context.WithoutCancel(ctx), attempt, spool,
			"cancelled", "CANCELLATION_REQUESTED", false, uuid.Nil, [32]byte{})
	}
	cause := context.Cause(ctx)
	if cause == nil {
		cause = context.Canceled
	}
	return RenderExecutionResult{State: RenderExecutionLeaseLost}, cause
}

func (worker *Worker) authorize(ctx context.Context, pin Pin, actorID uuid.UUID) error {
	return worker.authorizer.Authorize(ctx, AuthorizationRequest{TenantID: pin.TenantID,
		ActorID: actorID, DashboardID: pin.DashboardID,
		TokenVersion: pin.TokenVersion, AuthorizationVersion: pin.AuthorizationVersion,
		AuthorizationClock: pin.AuthorizationClock, Permission: string(PermissionExport)})
}
