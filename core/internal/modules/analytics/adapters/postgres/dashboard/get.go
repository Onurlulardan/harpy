package dashboard

import (
	"context"

	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	"github.com/google/uuid"
)

func (storage *store) Get(ctx context.Context, id uuid.UUID) (dashboardapp.Snapshot, error) {
	if !storage.usable(ctx) {
		return dashboardapp.Snapshot{}, dashboardapp.ErrUnavailable
	}
	if !validUUIDv7(id) {
		return dashboardapp.Snapshot{}, dashboardapp.ErrInvalid
	}
	return scanSnapshot(storage.queryRow(ctx, getSQL, id), storage.notFound)
}
