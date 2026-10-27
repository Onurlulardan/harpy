package purge

import (
	"context"
	"time"

	purgeapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/purge"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) PersistCompletion(ctx context.Context,
	mutation purgeapp.Mutation,
) (purgeapp.Result, error) {
	if !storage.usable(ctx) || mutation.TenantID != storage.tenantID ||
		!mutation.ItemID.UUIDv7() || mutation.State != domain.ItemPurging ||
		!mutation.OperationID.UUIDv7() || !mutation.AuditID.UUIDv7() ||
		mutation.OccurredAt.IsZero() {
		return purgeapp.Result{}, purgeapp.ErrInvalid
	}
	var item string
	var purgedAt time.Time
	err := storage.queryRow(ctx, finishSQL, mutation.ItemID.String(),
		mutation.OccurredAt, mutation.OperationID.String(),
		mutation.AuditID.String()).Scan(&item, &purgedAt)
	if err != nil {
		if storage.notFound(err) {
			return purgeapp.Result{ItemID: mutation.ItemID, State: domain.ItemPurging,
				PendingBlobs: len(mutation.Versions)}, nil
		}
		return purgeapp.Result{}, purgeapp.ErrUnavailable
	}
	if item != mutation.ItemID.String() || purgedAt.IsZero() {
		return purgeapp.Result{}, purgeapp.ErrUnavailable
	}
	return purgeapp.Result{ItemID: mutation.ItemID, State: domain.ItemPurged,
		PurgedAt: purgedAt}, nil
}
