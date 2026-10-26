package publication

import (
	"context"

	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type TransitionRequest struct {
	ReleaseID        uuid.UUID
	ExpectedRevision uint64
	IdempotencyKey   string
}

func (application *Application) Submit(
	ctx context.Context, scope platformtx.TenantScope, request TransitionRequest,
) (Snapshot, error) {
	if !validUUIDv7(request.ReleaseID) || request.ExpectedRevision == 0 ||
		!tokenPattern.MatchString(request.IdempotencyKey) {
		return Snapshot{}, ErrInvalid
	}
	store, identity, current, timing, err := application.transitionStore(ctx, scope,
		OperationSubmit, request.ReleaseID, false)
	if err != nil {
		return Snapshot{}, err
	}
	if current.State != analyticspublication.StateDraft &&
		current.State != analyticspublication.StatePendingApproval {
		return Snapshot{}, ErrConflict
	}
	return application.submitBound(ctx, scope, store, identity, current, timing, request)
}

func (application *Application) submitBound(
	ctx context.Context,
	scope platformtx.TenantScope,
	store Store,
	identity ScopeIdentity,
	current Snapshot,
	timing timeBinding,
	request TransitionRequest,
) (Snapshot, error) {
	operationID, err := application.nextID()
	if err != nil {
		return Snapshot{}, err
	}
	requestHash := transitionDigest(OperationSubmit, identity, request)
	decision, err := application.decideMutation(ctx, scope, identity,
		current.ResourceKind, operationID, request.ReleaseID,
		request.IdempotencyKey, requestHash)
	if err != nil {
		return Snapshot{}, err
	}
	result, replay, err := store.Submit(ctx, TransitionMutation{OperationID: decision.OperationID,
		ActorID: identity.ActorID, ReleaseID: decision.OwnerFactID,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: requestHash,
		ExpectedRevision: request.ExpectedRevision, StepUpAt: timing.stepUp, OccurredAt: timing.now})
	if err != nil {
		return Snapshot{}, err
	}
	if validateSnapshot(result) != nil || result.ReleaseID != decision.OwnerFactID ||
		result.State != analyticspublication.StatePendingApproval || replay != decision.Replay ||
		(!replay && result.Revision != request.ExpectedRevision+1) {
		return Snapshot{}, ErrUnavailable
	}
	return result, nil
}
