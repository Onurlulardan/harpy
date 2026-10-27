package postgres

import (
	"context"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func (reader *profileReader) HistoricalSchema(
	ctx context.Context,
	schemaID uuid.UUID,
) (profileapp.Schema, error) {
	if !reader.usable(ctx) || schemaID == uuid.Nil ||
		schemaID.Variant() != uuid.RFC4122 {
		return profileapp.Schema{}, errProfileSnapshotUnavailable
	}
	schema, err := reader.schemaSnapshot(ctx, profileHistoricalSchemaSQL, schemaID)
	if err != nil {
		return profileapp.Schema{}, err
	}
	if schema.ID() != schemaID {
		return profileapp.Schema{}, errProfileSnapshotUnavailable
	}
	return schema, nil
}
