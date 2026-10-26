package postgres

import (
	"context"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
)

func (store *layoutStore) Resolve(
	ctx context.Context,
	resource layoutapi.ResourceKind,
) (layoutapp.PublishedPresentation, error) {
	if !store.queryUsable(ctx) {
		return layoutapp.PublishedPresentation{}, layoutUnavailable()
	}
	var document []byte
	if err := store.queryRow(
		ctx, resolveSQL, string(resource),
	).Scan(&document); err != nil {
		return layoutapp.PublishedPresentation{}, mapLayoutError(err)
	}
	return decodePublishedPresentation(document)
}
