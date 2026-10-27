package query

import (
	"context"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
)

func (storage *store) Publish(
	ctx context.Context, input queryapp.PublishMutation,
) (queryapp.Snapshot, bool, error) {
	if !storage.usable(ctx) {
		return queryapp.Snapshot{}, false, queryapp.ErrUnavailable
	}
	if !storage.command || input.ActorID != storage.actorID {
		return queryapp.Snapshot{}, false, queryapp.ErrForbidden
	}
	if !validMutation(input.OperationID, input.ActorID, input.IdempotencyKey, input.OccurredAt) ||
		!validUUIDv7(input.QueryID) || !validUUIDv7(input.DraftVersionID) ||
		!validUUIDv7(input.NextDraftVersionID) || input.DraftVersionID == input.NextDraftVersionID ||
		input.ExpectedRootRevision == 0 || input.ExpectedDraftRevision == 0 ||
		!input.CachePolicy.Valid() || !validDocuments(input.Contract.Query,
		input.Contract.Parameters, input.Contract.Output) {
		return queryapp.Snapshot{}, false, queryapp.ErrInvalid
	}
	dependencies, err := encodeDependencies(input.Dependencies)
	if err != nil {
		return queryapp.Snapshot{}, false, err
	}
	statement := publishSQL
	if storage.operation == queryapp.OperationRevalidate {
		statement = revalidateSQL
	} else if storage.operation != queryapp.OperationPublish {
		return queryapp.Snapshot{}, false, queryapp.ErrForbidden
	}
	digest := input.RequestSHA256
	manifest := input.ManifestSHA256
	return scanMutation(storage.queryRow(ctx, statement,
		input.OperationID, input.ActorID, input.IdempotencyKey, digest[:], input.QueryID,
		input.DraftVersionID, input.NextDraftVersionID, input.ExpectedRootRevision,
		input.ExpectedDraftRevision, input.Contract.Query.CanonicalJSON(),
		input.Contract.Parameters.CanonicalJSON(), input.Contract.Output.CanonicalJSON(),
		dependencies, manifest[:], string(input.CachePolicy), input.OccurredAt), storage.notFound)
}
