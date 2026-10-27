package api

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const (
	maximumCatalogDefaultSelect     = 50
	maximumCatalogSearchFields      = 16
	minimumUUIDCursorCanonicalBytes = 36
)

var errInvalidFieldCatalog = errors.New("REST Query field catalog is invalid")

type FieldCatalogInput struct {
	Reference     CatalogReference
	Storage       ResourceStorage
	Capabilities  CatalogCapabilities
	Fields        []FieldDefinition
	Relations     []RelationDefinition
	DefaultSelect []uuid.UUID
	StableID      uuid.UUID
	SearchFields  []uuid.UUID
}

type FieldCatalog struct {
	reference     CatalogReference
	storage       ResourceStorage
	capabilities  CatalogCapabilities
	fields        []FieldDefinition
	relations     []RelationDefinition
	defaultSelect []uuid.UUID
	stableID      uuid.UUID
	searchFields  []uuid.UUID
	validated     bool
}

func NewFieldCatalog(input FieldCatalogInput) (FieldCatalog, error) {
	catalog := FieldCatalog{
		reference: input.Reference, storage: input.Storage,
		capabilities:  input.Capabilities,
		fields:        cloneFieldDefinitions(input.Fields),
		relations:     append(make([]RelationDefinition, 0, len(input.Relations)), input.Relations...),
		defaultSelect: append(make([]uuid.UUID, 0, len(input.DefaultSelect)), input.DefaultSelect...),
		stableID:      input.StableID,
		searchFields:  append(make([]uuid.UUID, 0, len(input.SearchFields)), input.SearchFields...),
		validated:     true,
	}
	if err := catalog.Validate(); err != nil {
		return FieldCatalog{}, err
	}
	return catalog, nil
}

func (catalog FieldCatalog) Validate() error {
	if !catalog.validated || catalog.reference.Validate() != nil ||
		catalog.storage.Validate() != nil || catalog.capabilities.Validate() != nil ||
		catalog.fields == nil || catalog.relations == nil ||
		catalog.defaultSelect == nil || catalog.searchFields == nil {
		return errInvalidFieldCatalog
	}
	fields, valid := catalogMemberFields(catalog)
	if !valid || !validCatalogProfiles(catalog, fields) {
		return errInvalidFieldCatalog
	}
	return nil
}

func (FieldCatalog) String() string   { return queryValueRedaction }
func (FieldCatalog) GoString() string { return queryValueRedaction }
func (FieldCatalog) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (FieldCatalog) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }

func (FieldCatalogInput) String() string   { return queryValueRedaction }
func (FieldCatalogInput) GoString() string { return queryValueRedaction }
func (FieldCatalogInput) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (FieldCatalogInput) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }

func cloneFieldDefinitions(fields []FieldDefinition) []FieldDefinition {
	result := make([]FieldDefinition, len(fields))
	for index, field := range fields {
		result[index] = field.clone()
	}
	return result
}

func (catalog FieldCatalog) clone() FieldCatalog {
	result := catalog
	result.fields = cloneFieldDefinitions(catalog.fields)
	result.relations = append([]RelationDefinition{}, catalog.relations...)
	result.defaultSelect = append([]uuid.UUID{}, catalog.defaultSelect...)
	result.searchFields = append([]uuid.UUID{}, catalog.searchFields...)
	return result
}
