package postgres

import (
	"context"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

func (reader *profileReader) schemaSnapshot(
	ctx context.Context,
	statement string,
	arguments ...any,
) (profileapp.Schema, error) {
	if !reader.usable(ctx) || statement == "" {
		return profileapp.Schema{}, errProfileSnapshotUnavailable
	}
	var document []byte
	if err := reader.queryRow(ctx, statement, arguments...).Scan(&document); err != nil {
		return profileapp.Schema{}, errProfileSnapshotUnavailable
	}
	if len(document) == 0 {
		return profileapp.Schema{}, profileapp.ErrSchemaNotFound
	}
	return decodeProfileSchemaSnapshot(document)
}

func decodeProfileSchemaSnapshot(document []byte) (profileapp.Schema, error) {
	var wire profileSchemaSnapshotWire
	if decodeProfileJSON(document, &wire) != nil || wire.Fields == nil {
		return profileapp.Schema{}, errProfileSnapshotUnavailable
	}
	schemaID, err := decodeProfileUUID(wire.SchemaID)
	if err != nil {
		return profileapp.Schema{}, errProfileSnapshotUnavailable
	}
	definition, err := decodeProfileSHA256(wire.DefinitionSHA256)
	if err != nil {
		return profileapp.Schema{}, errProfileSnapshotUnavailable
	}
	fields := make([]profileapp.Field, 0, len(wire.Fields))
	for _, fieldWire := range wire.Fields {
		field, fieldErr := fieldWire.field(profileapp.ResourceKind(wire.Resource))
		if fieldErr != nil {
			return profileapp.Schema{}, errProfileSnapshotUnavailable
		}
		fields = append(fields, field)
	}
	schema, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: schemaID, Resource: profileapp.ResourceKind(wire.Resource),
		Version: wire.Version, State: profileapp.SchemaState(wire.State),
		Fields: fields, ExpectedDefinitionSHA256: definition,
	})
	if err != nil {
		return profileapp.Schema{}, errProfileSnapshotUnavailable
	}
	return schema, nil
}
