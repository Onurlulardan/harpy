package link

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Release(ctx context.Context,
	scope platformtx.TenantScope, request ReleaseRequest,
) (domain.Link, error) {
	if application == nil || scope.OperationID() != OperationRelease ||
		!request.LinkID.UUIDv7() || !foundation.ValidIdempotency(request.IdempotencyKey) {
		return domain.Link{}, ErrInvalid
	}
	if err := application.authority.Authorize(ctx, scope, OperationRelease,
		request.Source); err != nil {
		return domain.Link{}, ErrForbidden
	}
	store, identity, err := application.binder.Bind(ctx, scope)
	if err != nil || store == nil || !identity.Valid() {
		return domain.Link{}, ErrUnavailable
	}
	_, operationID, auditID, err := application.identifiers()
	if err != nil {
		return domain.Link{}, err
	}
	now := application.clock.Now().UTC()
	digest := foundation.Digest([]byte(OperationRelease), request.LinkID[:],
		request.Source.ID[:], request.Source.VersionID[:])
	result, _, err := store.Release(ctx, Mutation{Link: domain.Link{ID: request.LinkID,
		TenantID: identity.TenantID}, ActorID: identity.ActorID,
		OperationID: operationID, AuditID: auditID, Source: request.Source,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: digest,
		OccurredAt: now})
	return result, mapError(err)
}
