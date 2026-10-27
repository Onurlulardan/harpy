package query

import (
	"context"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	"github.com/google/uuid"
)

func (storage *store) Get(ctx context.Context, queryID uuid.UUID) (queryapp.Snapshot, error) {
	if !storage.usable(ctx) {
		return queryapp.Snapshot{}, queryapp.ErrUnavailable
	}
	if !validUUIDv7(queryID) {
		return queryapp.Snapshot{}, queryapp.ErrInvalid
	}
	return scanSnapshot(storage.queryRow(ctx, getSQL, queryID), storage.notFound)
}
