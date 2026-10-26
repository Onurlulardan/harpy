package postgres

import (
	"context"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func (store *layoutStore) List(
	ctx context.Context,
	resource layoutapi.ResourceKind,
) ([]layoutapi.Revision, error) {
	if !store.queryUsable(ctx) {
		return nil, layoutUnavailable()
	}
	var document []byte
	if err := store.queryRow(ctx, listSQL, string(resource)).Scan(&document); err != nil {
		return nil, mapLayoutError(err)
	}
	return decodeRevisionList(document)
}
