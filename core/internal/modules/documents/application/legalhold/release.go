package legalhold

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Release(ctx context.Context,
	scope platformtx.TenantScope, request ReleaseRequest,
) (domain.LegalHold, error) {
	if application == nil || scope.OperationID() != OperationRelease ||
		!request.HoldID.UUIDv7() || !foundation.ValidIdempotency(request.IdempotencyKey) {
		return domain.LegalHold{}, ErrInvalid
	}
	if err := application.authority.Authorize(ctx, scope, OperationRelease,
		request.Source); err != nil {
		return domain.LegalHold{}, ErrForbidden
	}
	store, identity, err := application.binder.Bind(ctx, scope)
	if err != nil || store == nil || !identity.Valid() {
		return domain.LegalHold{}, ErrUnavailable
	}
	_, operationID, auditID, err := application.identifiers()
	if err != nil {
		return domain.LegalHold{}, err
	}
	now := application.clock.Now().UTC()
	digest := foundation.Digest([]byte(OperationRelease), request.HoldID[:],
		[]byte(request.Source.Owner), []byte(request.Source.CaseID))
	result, _, err := store.Release(ctx, Mutation{Hold: domain.LegalHold{
		ID: request.HoldID, TenantID: identity.TenantID}, ActorID: identity.ActorID,
		OperationID: operationID, AuditID: auditID, Source: request.Source,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: digest, OccurredAt: now})
	return result, mapError(err)
}
