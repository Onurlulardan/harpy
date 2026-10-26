package reports

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func buildCatalogEntry(spec resourceSpec) (catalogEntry, error) {
	hash, hashErr := catalogHash(spec)
	reference, referenceErr := api.NewCatalogReference(
		Owner, spec.ResourceID, spec.VersionID, hash,
	)
	storage, storageErr := resourceStorage(spec)
	capabilities, capabilitiesErr := api.NewCatalogCapabilities(!spec.Aggregate)
	fields, defaults, fieldsErr := buildFields(spec)
	stableID, stableErr := uuid.Parse(spec.StableID)
	if hashErr != nil || referenceErr != nil || storageErr != nil ||
		capabilitiesErr != nil || fieldsErr != nil || stableErr != nil {
		return catalogEntry{}, ErrContract
	}
	root, err := api.NewFieldCatalog(api.FieldCatalogInput{
		Reference: reference, Storage: storage, Capabilities: capabilities,
		Fields: fields, Relations: []api.RelationDefinition{},
		DefaultSelect: defaults, StableID: stableID, SearchFields: []uuid.UUID{},
	})
	if err != nil {
		return catalogEntry{}, ErrContract
	}
	set, err := api.NewCatalogSet(reference, []api.FieldCatalog{root})
	if err != nil {
		return catalogEntry{}, ErrContract
	}
	return catalogEntry{spec: spec, reference: reference, set: set}, nil
}

func buildFields(spec resourceSpec) (
	[]api.FieldDefinition, []uuid.UUID, error,
) {
	fields := make([]api.FieldDefinition, len(spec.Fields))
	defaults := make([]uuid.UUID, 0, len(spec.Fields))
	for index, fieldSpec := range spec.Fields {
		field, err := buildField(fieldSpec)
		if err != nil {
			return nil, nil, ErrContract
		}
		identifier, err := field.ID()
		if err != nil {
			return nil, nil, ErrContract
		}
		fields[index] = field
		if fieldSpec.Default {
			defaults = append(defaults, identifier)
		}
	}
	return fields, defaults, nil
}
