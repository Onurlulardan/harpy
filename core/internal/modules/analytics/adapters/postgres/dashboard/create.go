package dashboard

import (
	"context"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
)

func (storage *store) Create(ctx context.Context, input dashboardapp.CreateMutation) (
	dashboardapp.Snapshot, bool, error,
) {
	if !storage.usable(ctx) {
		return dashboardapp.Snapshot{}, false, dashboardapp.ErrUnavailable
	}
	if !storage.command || input.ActorID != storage.actorID {
		return dashboardapp.Snapshot{}, false, dashboardapp.ErrForbidden
	}
	if !validMutation(input.OperationID, input.ActorID, input.IdempotencyKey, input.OccurredAt) ||
		!validUUIDv7(input.DashboardID) || !validUUIDv7(input.VersionID) ||
		!namePattern.MatchString(input.Name) || !input.Visibility.Valid() || !input.Layout.Valid() {
		return dashboardapp.Snapshot{}, false, dashboardapp.ErrInvalid
	}
	widgets, err := encodeWidgets(input.Widgets)
	if err != nil {
		return dashboardapp.Snapshot{}, false, err
	}
	digest := input.RequestSHA256
	return scanMutation(storage.queryRow(ctx, createSQL, input.OperationID, input.ActorID,
		input.IdempotencyKey, digest[:], input.DashboardID, input.VersionID, input.Name,
		string(input.Visibility), input.Layout.CanonicalJSON(), widgets, input.OccurredAt), storage.notFound)
}
