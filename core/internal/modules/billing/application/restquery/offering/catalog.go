package offering

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type Catalog struct {
	reference api.CatalogReference
	set       api.CatalogSet
	valid     bool
}

func NewCatalog() (*Catalog, error) {
	hash, hashErr := catalogHash()
	reference, referenceErr := api.NewCatalogReference(
		Owner, resourceID, publishedVersionID, hash,
	)
	storage, storageErr := resourceStorage()
	capabilities, capabilitiesErr := api.NewCatalogCapabilities(true)
	fields, defaults, searches, fieldsErr := buildFields()
	stableID, stableErr := uuid.Parse(fieldID)
	if hashErr != nil || referenceErr != nil || storageErr != nil ||
		capabilitiesErr != nil || fieldsErr != nil || stableErr != nil {
		return nil, ErrContract
	}
	root, err := api.NewFieldCatalog(api.FieldCatalogInput{
		Reference: reference, Storage: storage, Capabilities: capabilities,
		Fields: fields, Relations: []api.RelationDefinition{},
		DefaultSelect: defaults, StableID: stableID, SearchFields: searches,
	})
	if err != nil {
		return nil, ErrContract
	}
	set, err := api.NewCatalogSet(reference, []api.FieldCatalog{root})
	if err != nil {
		return nil, ErrContract
	}
	catalog := &Catalog{reference: reference, set: set, valid: true}
	if catalog.Validate() != nil {
		return nil, ErrContract
	}
	return catalog, nil
}

func buildFields() (
	[]api.FieldDefinition, []uuid.UUID, []uuid.UUID, error,
) {
	fields := make([]api.FieldDefinition, len(offeringFields))
	defaults := make([]uuid.UUID, 0, len(offeringFields))
	searches := make([]uuid.UUID, 0, 1)
	for index, spec := range offeringFields {
		field, err := buildField(spec)
		if err != nil {
			return nil, nil, nil, ErrContract
		}
		identifier, err := field.ID()
		if err != nil {
			return nil, nil, nil, ErrContract
		}
		fields[index] = field
		if spec.Default {
			defaults = append(defaults, identifier)
		}
		if spec.Search {
			searches = append(searches, identifier)
		}
	}
	return fields, defaults, searches, nil
}

func (catalog *Catalog) Validate() error {
	if catalog == nil || !catalog.valid || catalog.reference.Validate() != nil ||
		catalog.set.Validate() != nil {
		return ErrContract
	}
	root, err := catalog.set.Root()
	if err != nil || !sameReference(root, catalog.reference) {
		return ErrContract
	}
	return nil
}

func (catalog *Catalog) Catalog(
	ctx context.Context,
	scope transaction.GlobalScope,
	resource string,
) (api.CatalogSet, error) {
	if catalog.Validate() != nil || ctx == nil || ctx.Err() != nil ||
		resource != ResourceName || transaction.NetworkAllowed(ctx) ||
		!transaction.ContextOwnsGlobalScope(ctx, scope) ||
		transaction.RequireGlobalOwner(scope, Owner) != nil {
		return api.CatalogSet{}, ErrContract
	}
	return catalog.set, nil
}
