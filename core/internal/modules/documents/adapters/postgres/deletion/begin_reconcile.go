package deletion

import (
	"context"

	deletionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/deletion"
)

func (storage *store) BeginReconcile(ctx context.Context, task deletionapp.Task) error {
	if !storage.usable(ctx) || task.TenantID != storage.tenantID ||
		!task.ID.UUIDv7() || !task.LeaseOwnerID.UUIDv7() || task.LeaseGeneration == 0 {
		return deletionapp.ErrInvalid
	}
	var changed bool
	err := storage.queryRow(ctx, beginReconcileSQL, task.ID.String(),
		task.LeaseOwnerID.String(), task.LeaseGeneration).Scan(&changed)
	if err != nil {
		if storage.notFound(err) {
			return deletionapp.ErrConflict
		}
		return deletionapp.ErrUnavailable
	}
	if !changed {
		return deletionapp.ErrConflict
	}
	return nil
}
