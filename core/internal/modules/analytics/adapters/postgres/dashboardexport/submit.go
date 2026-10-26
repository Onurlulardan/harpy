package dashboardexport

import (
	"context"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

func (storage *store) Submit(
	ctx context.Context, mutation exportapp.SubmitMutation,
) (exportapp.Snapshot, bool, error) {
	if !storage.usable(ctx, true) {
		return exportapp.Snapshot{}, false, exportapp.ErrUnavailable
	}
	return scanSubmit(storage.queryRow(ctx, submitSQL,
		mutation.OperationID, mutation.ActorID, mutation.IdempotencyKey,
		mutation.RequestSHA256[:], mutation.ExportID, mutation.JobID,
		mutation.ExpiryJobID, mutation.DashboardID, string(mutation.Format), mutation.Parameters,
		mutation.Auth.TokenVersion, mutation.Auth.AuthorizationVersion,
		mutation.Auth.AuthorizationClock, mutation.MaximumRows,
		mutation.MaximumPlaintextBytes, mutation.OccurredAt, mutation.ExpiresAt), storage.notFound)
}
