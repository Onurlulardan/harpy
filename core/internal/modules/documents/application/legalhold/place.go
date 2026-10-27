package legalhold

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Place(ctx context.Context,
	scope platformtx.TenantScope, request PlaceRequest,
) (domain.LegalHold, error) {
	if application == nil || scope.OperationID() != OperationPlace ||
		!request.ItemID.UUIDv7() || request.EvidenceSHA256 == [32]byte{} ||
		!foundation.ValidIdempotency(request.IdempotencyKey) {
		return domain.LegalHold{}, ErrInvalid
	}
	if err := application.authority.Authorize(ctx, scope, OperationPlace,
		request.Source); err != nil {
		return domain.LegalHold{}, ErrForbidden
	}
	store, identity, err := application.binder.Bind(ctx, scope)
	if err != nil || store == nil || !identity.Valid() {
		return domain.LegalHold{}, ErrUnavailable
	}
	holdID, operationID, auditID, err := application.identifiers()
	if err != nil {
		return domain.LegalHold{}, err
	}
	now := application.clock.Now().UTC()
	hold, err := domain.NewLegalHold(holdID, identity.TenantID, request.ItemID,
		request.Source.Owner, request.Source.CaseID, request.EvidenceSHA256, now)
	if err != nil {
		return domain.LegalHold{}, ErrInvalid
	}
	digest := foundation.Digest([]byte(OperationPlace), request.ItemID[:],
		[]byte(request.Source.Owner), []byte(request.Source.CaseID),
		request.EvidenceSHA256[:])
	result, _, err := store.Place(ctx, Mutation{Hold: hold, ActorID: identity.ActorID,
		OperationID: operationID, AuditID: auditID, Source: request.Source,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: digest, OccurredAt: now})
	return result, mapError(err)
}
