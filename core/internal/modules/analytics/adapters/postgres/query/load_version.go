package query

import (
	"context"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	"github.com/google/uuid"
)

func (storage *store) LoadVersion(
	ctx context.Context, queryID, versionID uuid.UUID,
) (queryapp.Snapshot, error) {
	if !storage.usable(ctx) {
		return queryapp.Snapshot{}, queryapp.ErrUnavailable
	}
	if !validUUIDv7(queryID) || !validUUIDv7(versionID) {
		return queryapp.Snapshot{}, queryapp.ErrInvalid
	}
	return scanSnapshot(storage.queryRow(ctx, getVersionSQL, queryID, versionID), storage.notFound)
}
