package deletion

import (
	"context"
	"time"

	deletionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/deletion"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) Lease(ctx context.Context, deletionID, ownerID domain.ID,
	now time.Time, lease time.Duration,
) (deletionapp.Task, error) {
	if !storage.usable(ctx) || !deletionID.UUIDv7() || !ownerID.UUIDv7() ||
		now.IsZero() || lease <= 0 || lease > time.Minute {
		return deletionapp.Task{}, deletionapp.ErrInvalid
	}
	return scanTask(storage.queryRow(ctx, claimSQL, deletionID.String(),
		ownerID.String(), now, now.Add(lease)), storage.tenantID, storage.notFound)
}
