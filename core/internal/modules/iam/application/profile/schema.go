package profile

import (
	"bytes"
	"crypto/sha256"
	"sort"

	"github.com/google/uuid"
)

type SchemaInput struct {
	ID                       uuid.UUID
	Resource                 ResourceKind
	Version                  int64
	State                    SchemaState
	Fields                   []Field
	ExpectedDefinitionSHA256 []byte
}

type Schema struct {
	id             uuid.UUID
	resource       ResourceKind
	version        int64
	state          SchemaState
	fields         []Field
	definitionJSON []byte
	definitionSHA  [sha256.Size]byte
	validated      bool
}

func NewPublishedSchema(input SchemaInput) (Schema, error) {
	if input.ID == uuid.Nil || input.ID.Variant() != uuid.RFC4122 ||
		!input.Resource.Valid() || input.Version < 1 || !input.State.Valid() ||
		input.Fields == nil {
		return Schema{}, ErrSchemaInvalid
	}
	fields := cloneAndSortFields(input.Fields)
	if !validSchemaFields(input.Resource, fields) {
		return Schema{}, ErrSchemaInvalid
	}
	definition, err := canonicalMarshal(buildSchemaDefinition(input, fields), MaximumSchemaBytes)
	if err != nil {
		return Schema{}, ErrSchemaInvalid
	}
	digest := canonicalDigest(definition)
	if input.ExpectedDefinitionSHA256 != nil &&
		(len(input.ExpectedDefinitionSHA256) != sha256.Size ||
			!bytes.Equal(input.ExpectedDefinitionSHA256, digest[:])) {
		return Schema{}, ErrSchemaInvalid
	}
	return Schema{
		id: input.ID, resource: input.Resource, version: input.Version,
		state: input.State, fields: fields, definitionJSON: definition,
		definitionSHA: digest, validated: true,
	}, nil
}

func cloneAndSortFields(fields []Field) []Field {
	result := make([]Field, len(fields))
	for index := range fields {
		result[index] = fields[index].clone()
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].id.String() < result[right].id.String()
	})
	return result
}

func validSchemaFields(resource ResourceKind, fields []Field) bool {
	ids := make(map[uuid.UUID]struct{}, len(fields))
	keys := make(map[string]struct{}, len(fields))
	active := 0
	for _, field := range fields {
		if !field.valid() || field.resource != resource {
			return false
		}
		if _, duplicate := ids[field.id]; duplicate {
			return false
		}
		if _, duplicate := keys[field.key]; duplicate {
			return false
		}
		ids[field.id], keys[field.key] = struct{}{}, struct{}{}
		if field.state == FieldActive {
			active++
		}
	}
	return active <= MaximumActiveFields
}
