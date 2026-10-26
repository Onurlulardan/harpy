package publication

import (
	"context"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
)

func (storage *store) Create(ctx context.Context, input publicationapp.CreateMutation) (
	publicationapp.Snapshot, bool, error,
) {
	if !storage.usable(ctx) {
		return publicationapp.Snapshot{}, false, publicationapp.ErrUnavailable
	}
	if !storage.command || input.ActorID != storage.actorID {
		return publicationapp.Snapshot{}, false, publicationapp.ErrForbidden
	}
	if !validMutation(input.OperationID, input.ActorID, input.IdempotencyKey, input.OccurredAt) ||
		!validUUIDv7(input.ReleaseID) || !validUUIDv7(input.ResourceVersionID) ||
		!input.ResourceKind.Valid() || !input.Audience.Valid() || !input.CachePolicy.Valid() ||
		input.ManifestSHA256 == [32]byte{} {
		return publicationapp.Snapshot{}, false, publicationapp.ErrInvalid
	}
	digest, manifest := input.RequestSHA256, input.ManifestSHA256
	return scanMutation(storage.queryRow(ctx, createSQL, input.OperationID, input.ActorID,
		input.IdempotencyKey, digest[:], input.ReleaseID, string(input.ResourceKind),
		input.ResourceVersionID, string(input.Audience), string(input.CachePolicy),
		manifest[:], input.OccurredAt), storage.notFound)
}
