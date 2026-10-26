package profile

import "github.com/google/uuid"

type SchemaChange struct {
	Schema                 Schema
	ExpectedActiveSchemaID uuid.UUID
}

func (change SchemaChange) Validate() error {
	definition := change.Schema.DefinitionJSON()
	if !stableID(change.Schema.ID()) || !change.Schema.Resource().Valid() ||
		change.Schema.Version() < 2 || change.Schema.State() != SchemaActive ||
		len(definition) < 2 || len(definition) > MaximumSchemaBytes ||
		!stableID(change.ExpectedActiveSchemaID) ||
		change.ExpectedActiveSchemaID == change.Schema.ID() {
		return ErrInvalidCommand
	}
	fields := change.Schema.Fields()
	if fields == nil || len(fields) > MaximumActiveFields {
		return ErrInvalidCommand
	}
	for _, field := range fields {
		if !stableID(field.ID()) || field.Resource() != change.Schema.Resource() ||
			!field.State().Valid() {
			return ErrInvalidCommand
		}
	}
	return nil
}
