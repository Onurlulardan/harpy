package profile

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"

	"github.com/google/uuid"
)

type Document struct {
	schemaID  uuid.UUID
	resource  ResourceKind
	values    map[uuid.UUID]json.RawMessage
	semantics map[uuid.UUID][sha256.Size]byte
	canonical []byte
	digest    [sha256.Size]byte
}

func (document Document) SchemaID() uuid.UUID       { return document.schemaID }
func (document Document) Resource() ResourceKind    { return document.resource }
func (document Document) CanonicalJSON() []byte     { return bytes.Clone(document.canonical) }
func (document Document) SHA256() [sha256.Size]byte { return document.digest }

func (document Document) Value(id uuid.UUID) ([]byte, bool) {
	value, found := document.values[id]
	return bytes.Clone(value), found
}

func (document Document) valid() bool {
	return document.schemaID != uuid.Nil && document.resource.Valid() &&
		document.values != nil && document.semantics != nil && len(document.canonical) >= 2
}

func schemaSemantics(schema Schema) map[uuid.UUID][sha256.Size]byte {
	result := make(map[uuid.UUID][sha256.Size]byte, len(schema.fields))
	for _, field := range schema.fields {
		result[field.id] = field.semanticSHA
	}
	return result
}

func cloneValues(values map[uuid.UUID]json.RawMessage) map[uuid.UUID]json.RawMessage {
	result := make(map[uuid.UUID]json.RawMessage, len(values))
	for id, value := range values {
		result[id] = bytes.Clone(value)
	}
	return result
}
