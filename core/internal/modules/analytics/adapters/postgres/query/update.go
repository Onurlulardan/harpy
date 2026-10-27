package query

import (
	"context"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
)

func (storage *store) Update(
	ctx context.Context, input queryapp.UpdateMutation,
) (queryapp.Snapshot, bool, error) {
	if !storage.usable(ctx) {
		return queryapp.Snapshot{}, false, queryapp.ErrUnavailable
	}
	if !storage.command || input.ActorID != storage.actorID {
		return queryapp.Snapshot{}, false, queryapp.ErrForbidden
	}
	if !validMutation(input.OperationID, input.ActorID, input.IdempotencyKey, input.OccurredAt) ||
		!validUUIDv7(input.QueryID) || input.ExpectedRootRevision == 0 ||
		input.ExpectedVersionRevision == 0 || !namePattern.MatchString(input.Name) ||
		!validDocuments(input.Query, input.Parameters, input.Output) {
		return queryapp.Snapshot{}, false, queryapp.ErrInvalid
	}
	digest := input.RequestSHA256
	return scanMutation(storage.queryRow(ctx, updateSQL, input.OperationID, input.ActorID,
		input.IdempotencyKey, digest[:], input.QueryID, input.ExpectedRootRevision,
		input.ExpectedVersionRevision, input.Name, input.Query.CanonicalJSON(),
		input.Parameters.CanonicalJSON(), input.Output.CanonicalJSON(), input.OccurredAt),
		storage.notFound)
}
