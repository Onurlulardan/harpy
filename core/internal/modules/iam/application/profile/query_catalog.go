package profile

import (
	"bytes"
	"crypto/sha256"

	"github.com/google/uuid"
)

type QueryCatalog struct {
	schemaID  uuid.UUID
	resource  ResourceKind
	version   int64
	hash      [sha256.Size]byte
	canonical []byte
	fields    []QueryField
	validated bool
}

func NewQueryCatalog(schema Schema) (QueryCatalog, error) {
	return newQueryCatalog(schema, false)
}

// NewRetainedQueryCatalog rebuilds the immutable query contract of an exact
// retained schema. Retired schemas remain readable for already-pinned exports;
// callers must still compare the resulting catalog digest to their stored pin.
func NewRetainedQueryCatalog(schema Schema) (QueryCatalog, error) {
	return newQueryCatalog(schema, true)
}

func newQueryCatalog(schema Schema, retained bool) (QueryCatalog, error) {
	if !schema.valid() || schema.state != SchemaActive &&
		(!retained || schema.state != SchemaRetired) {
		return QueryCatalog{}, ErrSchemaInvalid
	}
	fields := make([]QueryField, 0, len(schema.fields))
	for _, field := range schema.fields {
		if field.state == FieldRetired {
			continue
		}
		queryField, err := newQueryField(field)
		if err != nil {
			return QueryCatalog{}, ErrQueryCapabilityRejected
		}
		fields = append(fields, queryField)
	}
	canonical, err := canonicalMarshal(buildQueryDefinition(schema, fields), MaximumSchemaBytes)
	if err != nil {
		return QueryCatalog{}, ErrQueryCapabilityRejected
	}
	return QueryCatalog{
		schemaID: schema.id, resource: schema.resource, version: schema.version,
		hash: canonicalDigest(canonical), canonical: canonical, fields: fields,
		validated: true,
	}, nil
}

func (catalog QueryCatalog) Validate() error {
	if !catalog.validated || catalog.schemaID == uuid.Nil || !catalog.resource.Valid() ||
		catalog.version < 1 || len(catalog.canonical) == 0 || catalog.fields == nil {
		return ErrQueryCapabilityRejected
	}
	for _, field := range catalog.fields {
		if field.Validate() != nil {
			return ErrQueryCapabilityRejected
		}
	}
	return nil
}

func (catalog QueryCatalog) SchemaID() (uuid.UUID, error) {
	if err := catalog.Validate(); err != nil {
		return uuid.Nil, err
	}
	return catalog.schemaID, nil
}
func (catalog QueryCatalog) Resource() (ResourceKind, error) {
	if err := catalog.Validate(); err != nil {
		return "", err
	}
	return catalog.resource, nil
}
func (catalog QueryCatalog) Version() (int64, error) {
	if err := catalog.Validate(); err != nil {
		return 0, err
	}
	return catalog.version, nil
}
func (catalog QueryCatalog) Hash() ([sha256.Size]byte, error) {
	if err := catalog.Validate(); err != nil {
		return [sha256.Size]byte{}, err
	}
	return catalog.hash, nil
}
func (catalog QueryCatalog) CanonicalJSON() ([]byte, error) {
	if err := catalog.Validate(); err != nil {
		return nil, err
	}
	return bytes.Clone(catalog.canonical), nil
}
func (catalog QueryCatalog) Fields() ([]QueryField, error) {
	if err := catalog.Validate(); err != nil {
		return nil, err
	}
	result := make([]QueryField, len(catalog.fields))
	for index, field := range catalog.fields {
		result[index] = field
		result[index].operators, _ = field.Operators()
	}
	return result, nil
}

func (catalog QueryCatalog) Field(id uuid.UUID) (QueryField, error) {
	if catalog.Validate() != nil {
		return QueryField{}, ErrQueryCapabilityRejected
	}
	for _, field := range catalog.fields {
		if field.id == id {
			field.operators, _ = field.Operators()
			return field, nil
		}
	}
	return QueryField{}, ErrFieldUnavailable
}

func (catalog QueryCatalog) FieldByPublicName(name string) (QueryField, error) {
	if catalog.Validate() != nil {
		return QueryField{}, ErrQueryCapabilityRejected
	}
	for _, field := range catalog.fields {
		if field.publicName == name {
			field.operators, _ = field.Operators()
			return field, nil
		}
	}
	return QueryField{}, ErrFieldUnavailable
}
