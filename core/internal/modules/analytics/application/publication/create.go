package publication

import (
	"context"
	"fmt"

	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type CreateRequest struct {
	ResourceKind      ResourceKind
	ResourceVersionID uuid.UUID
	Audience          analyticspublication.Audience
	CachePolicy       analyticspublication.CachePolicy
	IdempotencyKey    string
}

func (request CreateRequest) IsPublic() bool {
	return request.Audience == analyticspublication.AudiencePublic
}

func (request CreateRequest) IsCapability() bool {
	return request.Audience == analyticspublication.AudienceCapability
}

func (application *Application) Create(
	ctx context.Context, scope platformtx.TenantScope, request CreateRequest,
) (Snapshot, error) {
	permission, ok := permissionForAudience(request.Audience, false)
	if !ok || !request.ResourceKind.Valid() || !validUUIDv7(request.ResourceVersionID) ||
		!request.CachePolicy.Valid() || !tokenPattern.MatchString(request.IdempotencyKey) {
		return Snapshot{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand, OperationCreate, permission)
	if err != nil {
		return Snapshot{}, err
	}
	manifest, err := application.resources.VerifyImmutable(ctx, scope,
		request.ResourceKind, request.ResourceVersionID)
	if err != nil || manifest == [32]byte{} {
		return Snapshot{}, ErrConflict
	}
	return application.createBound(ctx, scope, store, identity, request, manifest)
}

func (application *Application) createBound(
	ctx context.Context,
	scope platformtx.TenantScope,
	store Store,
	identity ScopeIdentity,
	request CreateRequest,
	manifest [32]byte,
) (Snapshot, error) {
	operationID, err := application.nextID()
	if err != nil {
		return Snapshot{}, err
	}
	releaseID, err := application.nextID()
	if err != nil || releaseID == operationID {
		return Snapshot{}, ErrUnavailable
	}
	now := application.clock.Now()
	if !validTime(now) {
		return Snapshot{}, ErrUnavailable
	}
	requestHash := requestDigest([]byte(OperationCreate), []byte(identity.TenantID.String()),
		[]byte(identity.ActorID.String()), []byte(request.ResourceKind),
		request.ResourceVersionID[:], []byte(request.Audience), []byte(request.CachePolicy), manifest[:])
	decision, err := application.decideMutation(ctx, scope, identity,
		request.ResourceKind, operationID, releaseID, request.IdempotencyKey, requestHash)
	if err != nil {
		return Snapshot{}, err
	}
	result, replay, err := store.Create(ctx, CreateMutation{OperationID: decision.OperationID,
		ActorID: identity.ActorID, ReleaseID: decision.OwnerFactID,
		ResourceVersionID: request.ResourceVersionID, IdempotencyKey: request.IdempotencyKey,
		RequestSHA256: requestHash, ResourceKind: request.ResourceKind,
		Audience: request.Audience, CachePolicy: request.CachePolicy,
		ManifestSHA256: manifest, OccurredAt: now})
	if err != nil {
		return Snapshot{}, fmt.Errorf("create analytics publication: %w", err)
	}
	if validateSnapshot(result) != nil || result.ReleaseID != decision.OwnerFactID ||
		result.ManifestSHA256 != manifest || replay != decision.Replay {
		return Snapshot{}, ErrUnavailable
	}
	return result, nil
}
