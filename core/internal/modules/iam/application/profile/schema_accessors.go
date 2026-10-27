package profile

import (
	"bytes"
	"crypto/sha256"

	"github.com/google/uuid"
)

func (schema Schema) ID() uuid.UUID                       { return schema.id }
func (schema Schema) Resource() ResourceKind              { return schema.resource }
func (schema Schema) Version() int64                      { return schema.version }
func (schema Schema) State() SchemaState                  { return schema.state }
func (schema Schema) DefinitionSHA256() [sha256.Size]byte { return schema.definitionSHA }
func (schema Schema) DefinitionJSON() []byte              { return bytes.Clone(schema.definitionJSON) }

func (schema Schema) Fields() []Field {
	return cloneAndSortFields(schema.fields)
}

func (schema Schema) Field(id uuid.UUID) (Field, bool) {
	for _, field := range schema.fields {
		if field.id == id {
			return field.clone(), true
		}
	}
	return Field{}, false
}

func (schema Schema) valid() bool {
	return schema.validated && schema.id != uuid.Nil && schema.resource.Valid() &&
		schema.version > 0 && schema.state.Valid() && schema.fields != nil &&
		len(schema.definitionJSON) > 0
}
