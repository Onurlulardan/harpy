package postgres

import (
	"context"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

func (reader *profileReader) ActiveQueryCatalog(
	ctx context.Context,
	resource profileapp.ResourceKind,
) (profileapp.QueryCatalog, error) {
	schema, err := reader.ActiveSchema(ctx, resource)
	if err != nil {
		return profileapp.QueryCatalog{}, errProfileSnapshotUnavailable
	}
	catalog, err := profileapp.NewQueryCatalog(schema)
	if err != nil {
		return profileapp.QueryCatalog{}, errProfileSnapshotUnavailable
	}
	return catalog, nil
}
