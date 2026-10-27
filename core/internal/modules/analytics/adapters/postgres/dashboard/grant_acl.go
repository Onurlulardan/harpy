package dashboard

import (
	"context"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
)

func (storage *store) GrantACL(ctx context.Context, input dashboardapp.ACLMutation) (
	dashboardapp.ACL, bool, error,
) {
	if !storage.usable(ctx) {
		return dashboardapp.ACL{}, false, dashboardapp.ErrUnavailable
	}
	if !storage.command || input.ActorID != storage.actorID {
		return dashboardapp.ACL{}, false, dashboardapp.ErrForbidden
	}
	if !validMutation(input.OperationID, input.ActorID, input.IdempotencyKey, input.OccurredAt) ||
		!validUUIDv7(input.DashboardID) || input.PrincipalID == [16]byte{} ||
		(input.PrincipalKind != "user" && input.PrincipalKind != "role") ||
		(input.Level != "read" && input.Level != "edit" && input.Level != "publish") {
		return dashboardapp.ACL{}, false, dashboardapp.ErrInvalid
	}
	digest := input.RequestSHA256
	result := dashboardapp.ACL{}
	var replay bool
	err := storage.queryRow(ctx, grantACLSQL, input.OperationID, input.ActorID,
		input.IdempotencyKey, digest[:], input.DashboardID, input.PrincipalKind,
		input.PrincipalID, input.Level, input.ExpectedRevision, input.OccurredAt).Scan(
		&result.DashboardID, &result.PrincipalKind, &result.PrincipalID,
		&result.Level, &result.Revision, &replay)
	if err != nil {
		return dashboardapp.ACL{}, false, mapError(err, storage.notFound)
	}
	return result, replay, nil
}
