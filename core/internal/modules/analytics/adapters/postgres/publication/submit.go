package publication

import (
	"context"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
)

func (storage *store) Submit(ctx context.Context, input publicationapp.TransitionMutation) (
	publicationapp.Snapshot, bool, error,
) {
	return storage.transition(ctx, submitSQL, input, true)
}

func (storage *store) transition(ctx context.Context, statement string,
	input publicationapp.TransitionMutation, stepUp bool,
) (publicationapp.Snapshot, bool, error) {
	if !storage.usable(ctx) {
		return publicationapp.Snapshot{}, false, publicationapp.ErrUnavailable
	}
	if !storage.command || input.ActorID != storage.actorID {
		return publicationapp.Snapshot{}, false, publicationapp.ErrForbidden
	}
	if !validTransition(input.OperationID, input.ActorID, input.IdempotencyKey,
		input.RequestSHA256, input.ExpectedRevision, input.OccurredAt) ||
		!validUUIDv7(input.ReleaseID) || stepUp && (input.StepUpAt.IsZero() ||
		input.StepUpAt.After(input.OccurredAt) || input.OccurredAt.Sub(input.StepUpAt) > 10*60*1e9) {
		return publicationapp.Snapshot{}, false, publicationapp.ErrInvalid
	}
	digest := input.RequestSHA256
	if stepUp {
		return scanMutation(storage.queryRow(ctx, statement, input.OperationID, input.ActorID,
			input.IdempotencyKey, digest[:], input.ReleaseID, input.ExpectedRevision,
			input.StepUpAt, input.OccurredAt), storage.notFound)
	}
	return scanMutation(storage.queryRow(ctx, statement, input.OperationID, input.ActorID,
		input.IdempotencyKey, digest[:], input.ReleaseID, input.ExpectedRevision,
		input.OccurredAt), storage.notFound)
}
