package link

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Register(ctx context.Context,
	scope platformtx.TenantScope, request RegisterRequest,
) (domain.Link, error) {
	if application == nil || scope.OperationID() != OperationRegister ||
		!request.ItemID.UUIDv7() || !foundation.ValidIdempotency(request.IdempotencyKey) {
		return domain.Link{}, ErrInvalid
	}
	if err := application.authority.Authorize(ctx, scope, OperationRegister,
		request.Source); err != nil {
		return domain.Link{}, ErrForbidden
	}
	store, identity, err := application.binder.Bind(ctx, scope)
	if err != nil || store == nil || !identity.Valid() {
		return domain.Link{}, ErrUnavailable
	}
	linkID, operationID, auditID, err := application.identifiers()
	if err != nil {
		return domain.Link{}, err
	}
	now := application.clock.Now()
	value, err := domain.NewLink(linkID, identity.TenantID, request.ItemID,
		request.Source.ID, request.Source.VersionID, request.Source.Owner,
		request.Source.Kind, request.Usage, request.ReferenceEpoch, now)
	if err != nil {
		return domain.Link{}, ErrInvalid
	}
	digest := foundation.Digest([]byte(OperationRegister), request.ItemID[:],
		request.Source.ID[:], request.Source.VersionID[:], []byte(request.Source.Owner),
		[]byte(request.Source.Kind), []byte(request.Usage))
	result, _, err := store.Register(ctx, Mutation{Link: value, ActorID: identity.ActorID,
		OperationID: operationID, AuditID: auditID, Source: request.Source,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: digest,
		OccurredAt: now.UTC()})
	return result, mapError(err)
}

func (application *Application) identifiers() (domain.ID, domain.ID, domain.ID, error) {
	values := make([]domain.ID, 3)
	seen := map[domain.ID]struct{}{}
	for index := range values {
		value, err := application.ids.NewUUIDv7()
		_, duplicate := seen[value]
		if err != nil || !value.UUIDv7() || duplicate {
			return domain.ID{}, domain.ID{}, domain.ID{}, ErrUnavailable
		}
		values[index], seen[value] = value, struct{}{}
	}
	return values[0], values[1], values[2], nil
}
