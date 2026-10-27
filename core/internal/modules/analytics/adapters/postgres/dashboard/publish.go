package dashboard

import (
	"context"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
)

func (storage *store) Publish(ctx context.Context, input dashboardapp.PublishMutation) (
	dashboardapp.Snapshot, bool, error,
) {
	if !storage.usable(ctx) {
		return dashboardapp.Snapshot{}, false, dashboardapp.ErrUnavailable
	}
	if !storage.command || input.ActorID != storage.actorID {
		return dashboardapp.Snapshot{}, false, dashboardapp.ErrForbidden
	}
	if !validMutation(input.OperationID, input.ActorID, input.IdempotencyKey, input.OccurredAt) ||
		!validUUIDv7(input.DashboardID) || !validUUIDv7(input.DraftVersionID) ||
		!validUUIDv7(input.NextDraftVersionID) || input.DraftVersionID == input.NextDraftVersionID ||
		input.ExpectedRootRevision == 0 || input.ExpectedDraftRevision == 0 ||
		input.ManifestSHA256 == [32]byte{} {
		return dashboardapp.Snapshot{}, false, dashboardapp.ErrInvalid
	}
	digest := input.RequestSHA256
	manifest := input.ManifestSHA256
	return scanMutation(storage.queryRow(ctx, publishSQL, input.OperationID, input.ActorID,
		input.IdempotencyKey, digest[:], input.DashboardID, input.DraftVersionID,
		input.NextDraftVersionID, input.ExpectedRootRevision, input.ExpectedDraftRevision,
		manifest[:], input.OccurredAt), storage.notFound)
}
