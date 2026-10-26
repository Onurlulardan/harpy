package postgres

import (
	"context"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

func (reader *profileReader) ActiveSchema(
	ctx context.Context,
	resource profileapp.ResourceKind,
) (profileapp.Schema, error) {
	if !reader.usable(ctx) || !resource.Valid() {
		return profileapp.Schema{}, errProfileSnapshotUnavailable
	}
	schema, err := reader.schemaSnapshot(ctx, profileActiveCatalogSQL, string(resource))
	if err != nil {
		return profileapp.Schema{}, err
	}
	if schema.Resource() != resource ||
		schema.State() != profileapp.SchemaActive {
		return profileapp.Schema{}, errProfileSnapshotUnavailable
	}
	return schema, nil
}
