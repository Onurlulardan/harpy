package dashboardexport

import (
	"context"
	"time"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/google/uuid"
)

func (storage *store) DescribeContent(
	ctx context.Context, dashboardID, exportID uuid.UUID, observedAt time.Time,
) (exportapp.StoredContent, error) {
	if !storage.usable(ctx, false) {
		return exportapp.StoredContent{}, exportapp.ErrUnavailable
	}
	stored := contentRow{tenantID: storage.tenantID}
	err := storage.queryRow(ctx, contentSQL, storage.actorID, dashboardID,
		exportID, observedAt).Scan(stored.destinations()...)
	if err != nil {
		return exportapp.StoredContent{}, mapError(err, storage.notFound)
	}
	return stored.decode()
}
