package publication

import (
	"context"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	"github.com/google/uuid"
)

func (storage *store) Get(ctx context.Context, id uuid.UUID) (publicationapp.Snapshot, error) {
	if !storage.usable(ctx) {
		return publicationapp.Snapshot{}, publicationapp.ErrUnavailable
	}
	if !validUUIDv7(id) {
		return publicationapp.Snapshot{}, publicationapp.ErrInvalid
	}
	return scanSnapshot(storage.queryRow(ctx, getSQL, id), storage.notFound)
}
