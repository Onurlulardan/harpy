package publication

import (
	"context"

	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Revoke(
	ctx context.Context, scope platformtx.TenantScope, request TransitionRequest,
) (Snapshot, error) {
	if !validUUIDv7(request.ReleaseID) || request.ExpectedRevision == 0 ||
		!tokenPattern.MatchString(request.IdempotencyKey) {
		return Snapshot{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand,
		OperationRevoke, PermissionRevoke)
	if err != nil {
		return Snapshot{}, err
	}
	current, err := store.Get(ctx, request.ReleaseID)
	if err != nil || current.State != analyticspublication.StateActive &&
		current.State != analyticspublication.StateRevoked {
		return Snapshot{}, ErrConflict
	}
	operationID, err := application.nextID()
	if err != nil {
		return Snapshot{}, err
	}
	now := application.clock.Now()
	if !validTime(now) {
		return Snapshot{}, ErrUnavailable
	}
	requestHash := transitionDigest(OperationRevoke, identity, request)
	result, replay, err := store.Revoke(ctx, TransitionMutation{OperationID: operationID,
		ActorID: identity.ActorID, ReleaseID: request.ReleaseID,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: requestHash,
		ExpectedRevision: request.ExpectedRevision, OccurredAt: now})
	if err != nil {
		return Snapshot{}, err
	}
	if validateSnapshot(result) != nil || result.State != analyticspublication.StateRevoked ||
		(!replay && (result.Revision != request.ExpectedRevision+1 ||
			result.Generation <= current.Generation)) {
		return Snapshot{}, ErrUnavailable
	}
	if !replay && application.effects.Revoked(ctx, scope, TransitionEffect{
		OperationID: operationID, ActorID: identity.ActorID, TenantID: identity.TenantID,
		ReleaseID: result.ReleaseID, ResourceVersionID: result.ResourceVersionID,
		ResourceKind: result.ResourceKind, Audience: result.Audience,
		Generation: result.Generation, OccurredAt: now,
	}) != nil {
		return Snapshot{}, ErrUnavailable
	}
	return result, nil
}
