package dashboardexport

import (
	"context"
	"time"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/google/uuid"
)

func (storage *store) Get(
	ctx context.Context, dashboardID, exportID uuid.UUID, observedAt time.Time,
) (exportapp.Snapshot, error) {
	if !storage.usable(ctx, false) {
		return exportapp.Snapshot{}, exportapp.ErrUnavailable
	}
	return scanSnapshot(storage.queryRow(ctx, getSQL, storage.actorID,
		dashboardID, exportID, observedAt), storage.notFound)
}
