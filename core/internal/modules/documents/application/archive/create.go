package archive

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Create(ctx context.Context, scope platformtx.TenantScope,
	request CreateRequest,
) (Snapshot, error) {
	_, document, digest, ok := canonicalRequest(request)
	if application == nil || !ok || !foundation.ValidIdempotency(request.IdempotencyKey) ||
		scope.OperationID() != OperationCreate {
		return Snapshot{}, ErrInvalid
	}
	revision, err := application.authorizer.Authorize(ctx, scope, PermissionCreate)
	if err != nil || revision == 0 {
		return Snapshot{}, ErrForbidden
	}
	store, identity, err := application.binder.Bind(ctx, scope, Command)
	if err != nil || store == nil || !identity.Valid() {
		return Snapshot{}, ErrUnavailable
	}
	if replay, found, findErr := store.FindCreate(ctx, request.IdempotencyKey, digest); findErr != nil {
		return Snapshot{}, mapError(findErr)
	} else if found {
		return replay, nil
	}
	ids, err := buildArchiveIDs(application.ids, request.RetryOfArchiveID)
	if err != nil {
		return Snapshot{}, err
	}
	now := application.clock.Now().UTC()
	mutation := CreateMutation{Snapshot: Snapshot{ID: ids.archive, TenantID: identity.TenantID,
		ActorID: identity.ActorID, JobID: ids.job, RetryOfArchiveID: request.RetryOfArchiveID,
		BillingReservationID: ids.reservation, State: domain.ArchivePending,
		AuthorizationRevision: revision, RequestSHA256: digest,
		CreatedAt: now, UpdatedAt: now}, OperationID: ids.operation, AuditID: ids.audit,
		RequestDocument: document, IdempotencyKey: request.IdempotencyKey, OccurredAt: now}
	result, err := store.Create(ctx, mutation)
	if err != nil {
		return Snapshot{}, mapError(err)
	}
	if err = application.jobs.Enqueue(ctx, scope, ids.job, ids.archive,
		jobDocument(ids.archive)); err != nil {
		return Snapshot{}, ErrUnavailable
	}
	return result, nil
}

type generatedIDs struct{ archive, job, reservation, operation, audit domain.ID }

func buildArchiveIDs(source foundation.IDSource, retryOf domain.ID) (generatedIDs, error) {
	result := generatedIDs{}
	values := []*domain.ID{&result.archive, &result.job, &result.reservation,
		&result.operation, &result.audit}
	seen := map[domain.ID]struct{}{}
	for _, target := range values {
		value, err := source.NewUUIDv7()
		_, duplicate := seen[value]
		if err != nil || !value.UUIDv7() || duplicate || value == retryOf {
			return generatedIDs{}, ErrUnavailable
		}
		*target, seen[value] = value, struct{}{}
	}
	return result, nil
}
