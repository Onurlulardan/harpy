package dashboard

import (
	"context"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
)

func (storage *store) RevokeACL(ctx context.Context, input dashboardapp.ACLMutation) (bool, error) {
	if !storage.usable(ctx) {
		return false, dashboardapp.ErrUnavailable
	}
	if !storage.command || input.ActorID != storage.actorID {
		return false, dashboardapp.ErrForbidden
	}
	if !validMutation(input.OperationID, input.ActorID, input.IdempotencyKey, input.OccurredAt) ||
		!validUUIDv7(input.DashboardID) || input.PrincipalID == [16]byte{} ||
		(input.PrincipalKind != "user" && input.PrincipalKind != "role") || input.ExpectedRevision == 0 {
		return false, dashboardapp.ErrInvalid
	}
	digest := input.RequestSHA256
	var replay bool
	err := storage.queryRow(ctx, revokeACLSQL, input.OperationID, input.ActorID,
		input.IdempotencyKey, digest[:], input.DashboardID, input.PrincipalKind,
		input.PrincipalID, input.ExpectedRevision, input.OccurredAt).Scan(&replay)
	return replay, mapError(err, storage.notFound)
}
