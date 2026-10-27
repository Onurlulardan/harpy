package dashboardexport

import (
	"context"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

func (storage *store) RequestDelete(
	ctx context.Context, mutation exportapp.DeleteMutation,
) (exportapp.Snapshot, bool, bool, error) {
	if !storage.usable(ctx, true) {
		return exportapp.Snapshot{}, false, false, exportapp.ErrUnavailable
	}
	return scanDelete(storage.queryRow(ctx, deleteSQL,
		mutation.OperationID, mutation.DeleteJobID, mutation.ActorID,
		mutation.IdempotencyKey, mutation.RequestSHA256[:], mutation.DashboardID,
		mutation.ExportID, mutation.OccurredAt), storage.notFound)
}
