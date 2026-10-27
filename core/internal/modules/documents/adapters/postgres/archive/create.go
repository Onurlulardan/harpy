package archive

import (
	"context"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
)

func (storage *store) Create(ctx context.Context,
	mutation archiveapp.CreateMutation,
) (archiveapp.Snapshot, error) {
	if !storage.usable(ctx) || storage.job || !storage.command ||
		mutation.ActorID != storage.actorID || mutation.TenantID != storage.tenantID ||
		!mutation.ID.UUIDv7() || !mutation.JobID.UUIDv7() ||
		!mutation.BillingReservationID.UUIDv7() || !mutation.OperationID.UUIDv7() ||
		!mutation.AuditID.UUIDv7() || len(mutation.RequestDocument) < 2 ||
		mutation.RequestSHA256 == [32]byte{} || mutation.AuthorizationRevision == 0 {
		return archiveapp.Snapshot{}, archiveapp.ErrInvalid
	}
	var retry any
	if mutation.RetryOfArchiveID.Valid() {
		retry = mutation.RetryOfArchiveID.String()
	}
	var id string
	err := storage.queryRow(ctx, createSQL, mutation.ID.String(), storage.actorID.String(),
		mutation.AuthorizationRevision, mutation.JobID.String(), retry,
		mutation.RequestSHA256[:], mutation.RequestDocument,
		mutation.BillingReservationID.String(), mutation.OccurredAt,
		mutation.OperationID.String(), mutation.IdempotencyKey,
		mutation.AuditID.String()).Scan(&id)
	if err != nil {
		if storage.notFound(err) {
			return archiveapp.Snapshot{}, archiveapp.ErrConflict
		}
		return archiveapp.Snapshot{}, archiveapp.ErrUnavailable
	}
	return storage.Get(ctx, mutation.ID)
}
