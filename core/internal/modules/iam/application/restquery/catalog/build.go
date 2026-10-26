package catalog

import (
	"encoding/hex"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

func Build(custom profile.QueryCatalog) (Published, error) {
	if custom.Validate() != nil {
		return Published{}, ErrContract
	}
	kind, kindErr := custom.Resource()
	schemaID, schemaErr := custom.SchemaID()
	profileHash, hashErr := custom.Hash()
	queryFields, fieldsErr := custom.Fields()
	resource, resourceErr := ResourceFor(kind)
	core, coreErr := coreFields(kind)
	if kindErr != nil || schemaErr != nil || hashErr != nil || fieldsErr != nil ||
		resourceErr != nil || coreErr != nil {
		return Published{}, ErrContract
	}
	definitions := make([]api.FieldDefinition, 0, len(core)+len(queryFields))
	classes := make(map[uuid.UUID]profile.Classification, len(core)+len(queryFields))
	required := make(map[uuid.UUID]bool, len(core)+len(queryFields))
	maximumBytes := make(map[uuid.UUID]int, len(core)+len(queryFields))
	for _, spec := range core {
		definition, err := newCoreField(spec)
		if err != nil {
			return Published{}, ErrContract
		}
		identifier, _ := definition.ID()
		if _, duplicate := classes[identifier]; duplicate {
			return Published{}, ErrContract
		}
		definitions = append(definitions, definition)
		classes[identifier] = spec.Classification
		required[identifier] = spec.CreateRequired
		maximumBytes[identifier] = spec.MaximumBytes
	}
	customDefinitions, customSearch, err := customFields(kind, queryFields, classes)
	if err != nil {
		return Published{}, ErrContract
	}
	for _, field := range queryFields {
		identifier, idErr := field.ID()
		createRequired, requiredErr := field.Required()
		maximum, maximumErr := field.MaximumBytes()
		if idErr != nil || requiredErr != nil || maximumErr != nil {
			return Published{}, ErrContract
		}
		if _, duplicate := required[identifier]; duplicate {
			return Published{}, ErrContract
		}
		required[identifier] = createRequired
		maximumBytes[identifier] = maximum
	}
	definitions = append(definitions, customDefinitions...)
	digest, err := finalDigest(resource, schemaID.String(), profileHash, core)
	if err != nil {
		return Published{}, ErrContract
	}
	return assemble(
		resource, schemaID.String(), digest, core, definitions, customSearch,
		classes, required, maximumBytes,
	)
}

func assemble(
	resource Resource,
	schemaID string,
	digest [32]byte,
	core []coreFieldSpec,
	fields []api.FieldDefinition,
	customSearch []uuid.UUID,
	classes map[uuid.UUID]profile.Classification,
	required map[uuid.UUID]bool,
	maximumBytes map[uuid.UUID]int,
) (Published, error) {
	hash, err := api.NewCatalogHash(hex.EncodeToString(digest[:]))
	if err != nil {
		return Published{}, ErrContract
	}
	reference, err := api.NewCatalogReference(
		Owner, resource.ID().String(), schemaID, hash,
	)
	table, tableErr := sqlidentifier.MintStatic(resource.Table())
	storage, storageErr := api.NewResourceStorage(table)
	capabilities, capabilitiesErr := api.NewCatalogCapabilities(true)
	stable, defaults, search, profilesErr := stableDefaultsAndSearch(core)
	if err != nil || tableErr != nil || storageErr != nil || capabilitiesErr != nil ||
		profilesErr != nil {
		return Published{}, ErrContract
	}
	search = append(search, customSearch...)
	fieldCatalog, err := api.NewFieldCatalog(api.FieldCatalogInput{
		Reference: reference, Storage: storage, Capabilities: capabilities,
		Fields: fields, Relations: []api.RelationDefinition{}, DefaultSelect: defaults,
		StableID: stable, SearchFields: search,
	})
	if err != nil {
		return Published{}, ErrContract
	}
	set, err := api.NewCatalogSet(reference, []api.FieldCatalog{fieldCatalog})
	if err != nil {
		return Published{}, ErrContract
	}
	published := Published{
		resource: resource, reference: reference, set: set,
		classifications: classes, createRequired: required,
		maximumBytes: maximumBytes, validated: true,
	}
	if published.Validate() != nil {
		return Published{}, ErrContract
	}
	return published, nil
}
