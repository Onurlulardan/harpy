package query

import (
	"context"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	"github.com/google/uuid"
)

func (storage *store) ResolveQueryID(
	ctx context.Context, versionID uuid.UUID,
) (uuid.UUID, error) {
	if !storage.usable(ctx) {
		return uuid.Nil, queryapp.ErrUnavailable
	}
	if !validUUIDv7(versionID) {
		return uuid.Nil, queryapp.ErrInvalid
	}
	var queryID uuid.UUID
	if err := storage.queryRow(ctx, resolveQueryIDSQL, versionID).Scan(&queryID); err != nil {
		return uuid.Nil, mapError(err, storage.notFound)
	}
	if !validUUIDv7(queryID) {
		return uuid.Nil, queryapp.ErrUnavailable
	}
	return queryID, nil
}
