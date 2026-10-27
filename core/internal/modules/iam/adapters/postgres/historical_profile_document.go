package postgres

import (
	"context"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func (reader *profileReader) HistoricalDocument(
	ctx context.Context,
	resource profileapp.ResourceKind,
	recordID uuid.UUID,
	normalizer profileapp.TextNormalizer,
) (profileapp.Document, error) {
	if !reader.usable(ctx) || !resource.Valid() || recordID == uuid.Nil ||
		recordID.Variant() != uuid.RFC4122 {
		return profileapp.Document{}, errProfileSnapshotUnavailable
	}
	var payload []byte
	if err := reader.queryRow(
		ctx, profileHistoricalDocumentSQL, string(resource), recordID,
	).Scan(&payload); err != nil {
		return profileapp.Document{}, errProfileSnapshotUnavailable
	}
	if len(payload) == 0 {
		return profileapp.Document{}, profileapp.ErrDocumentNotFound
	}
	wire, schemaID, digest, err := decodeProfileDocument(payload)
	if err != nil {
		return profileapp.Document{}, errProfileSnapshotUnavailable
	}
	schema, err := reader.schemaSnapshot(ctx, profileHistoricalSchemaSQL, schemaID)
	if err != nil {
		return profileapp.Document{}, err
	}
	if schema.ID() != schemaID || schema.Resource() != resource {
		return profileapp.Document{}, errProfileSnapshotUnavailable
	}
	document, err := profileapp.DecodeDocument(schema, wire.Values, digest, normalizer)
	if err != nil {
		return profileapp.Document{}, errProfileSnapshotUnavailable
	}
	return document, nil
}
