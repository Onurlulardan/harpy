package publication

import (
	"context"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
)

func (storage *store) Activate(ctx context.Context, input publicationapp.ActivateMutation) (
	publicationapp.Snapshot, bool, error,
) {
	transition := input.TransitionMutation
	if !storage.usable(ctx) {
		return publicationapp.Snapshot{}, false, publicationapp.ErrUnavailable
	}
	if !storage.command || transition.ActorID != storage.actorID {
		return publicationapp.Snapshot{}, false, publicationapp.ErrForbidden
	}
	if !validTransition(transition.OperationID, transition.ActorID, transition.IdempotencyKey,
		transition.RequestSHA256, transition.ExpectedRevision, transition.OccurredAt) ||
		!validUUIDv7(transition.ReleaseID) || transition.StepUpAt.IsZero() ||
		transition.StepUpAt.After(transition.OccurredAt) ||
		transition.OccurredAt.Sub(transition.StepUpAt) > 10*60*1e9 {
		return publicationapp.Snapshot{}, false, publicationapp.ErrInvalid
	}
	var publicID, capability any
	if input.PublicID != [16]byte{} {
		publicID = input.PublicID[:]
	}
	if input.CapabilitySHA256 != [32]byte{} {
		capability = input.CapabilitySHA256[:]
	}
	digest := transition.RequestSHA256
	return scanMutation(storage.queryRow(ctx, activateSQL, transition.OperationID,
		transition.ActorID, transition.IdempotencyKey, digest[:], transition.ReleaseID,
		transition.ExpectedRevision, transition.StepUpAt, publicID, capability,
		input.CapabilityExpiresAt, transition.OccurredAt), storage.notFound)
}
