package dashboard

import (
	"context"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
)

func (storage *store) SetDefault(ctx context.Context, input dashboardapp.DefaultMutation) (bool, error) {
	if !storage.usable(ctx) {
		return false, dashboardapp.ErrUnavailable
	}
	if !storage.command || input.ActorID != storage.actorID {
		return false, dashboardapp.ErrForbidden
	}
	if !validMutation(input.OperationID, input.ActorID, input.IdempotencyKey, input.OccurredAt) ||
		input.PrincipalID == [16]byte{} || !validUUIDv7(input.DashboardID) ||
		!surfacePattern.MatchString(input.Surface) {
		return false, dashboardapp.ErrInvalid
	}
	digest := input.RequestSHA256
	var replay bool
	err := storage.queryRow(ctx, setDefaultSQL, input.OperationID, input.ActorID,
		input.IdempotencyKey, digest[:], input.PrincipalID, input.DashboardID,
		input.Surface, input.OccurredAt).Scan(&replay)
	return replay, mapError(err, storage.notFound)
}
