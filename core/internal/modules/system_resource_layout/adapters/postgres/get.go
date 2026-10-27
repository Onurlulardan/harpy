package postgres

import (
	"context"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	"github.com/google/uuid"
)

func (store *layoutStore) Get(
	ctx context.Context,
	id uuid.UUID,
) (layoutapi.Revision, error) {
	if !store.queryUsable(ctx) {
		return layoutapi.Revision{}, layoutUnavailable()
	}
	var document []byte
	if err := store.queryRow(ctx, getSQL, id).Scan(&document); err != nil {
		return layoutapi.Revision{}, mapLayoutError(err)
	}
	return decodeRevision(document)
}
