package dashboard

import (
	"context"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
)

func (storage *store) Delete(ctx context.Context, input dashboardapp.DeleteMutation) (
	dashboardapp.Snapshot, bool, error,
) {
	if !storage.usable(ctx) {
		return dashboardapp.Snapshot{}, false, dashboardapp.ErrUnavailable
	}
	if !storage.command || input.ActorID != storage.actorID {
		return dashboardapp.Snapshot{}, false, dashboardapp.ErrForbidden
	}
	if !validMutation(input.OperationID, input.ActorID, input.IdempotencyKey, input.OccurredAt) ||
		!validUUIDv7(input.DashboardID) || input.ExpectedRootRevision == 0 {
		return dashboardapp.Snapshot{}, false, dashboardapp.ErrInvalid
	}
	digest := input.RequestSHA256
	return scanMutation(storage.queryRow(ctx, deleteSQL, input.OperationID, input.ActorID,
		input.IdempotencyKey, digest[:], input.DashboardID,
		input.ExpectedRootRevision, input.OccurredAt), storage.notFound)
}
