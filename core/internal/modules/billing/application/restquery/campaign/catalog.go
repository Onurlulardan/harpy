package campaign

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

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
	definitions, defaults, searches, fieldsErr := buildFields()
	stableID, stableErr := uuid.Parse(fieldID)
	if hashErr != nil || referenceErr != nil || storageErr != nil ||
		capabilitiesErr != nil || fieldsErr != nil || stableErr != nil {
		return nil, ErrContract
	}
	root, err := api.NewFieldCatalog(api.FieldCatalogInput{
		Reference: reference, Storage: storage, Capabilities: capabilities,
		Fields: definitions, Relations: []api.RelationDefinition{},
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

func buildFields() ([]api.FieldDefinition, []uuid.UUID, []uuid.UUID, error) {
	definitions := make([]api.FieldDefinition, len(fields))
	defaults := make([]uuid.UUID, 0, len(fields))
	searches := make([]uuid.UUID, 0, 1)
	for index, spec := range fields {
		field, err := buildField(spec)
		if err != nil {
			return nil, nil, nil, ErrContract
		}
		identifier, err := field.ID()
		if err != nil {
			return nil, nil, nil, ErrContract
		}
		definitions[index] = field
		if spec.Default {
			defaults = append(defaults, identifier)
		}
		if spec.Search {
			searches = append(searches, identifier)
		}
	}
	return definitions, defaults, searches, nil
}

func catalogHash() (api.CatalogHash, error) {
	document := struct {
		Version, Resource, Schema, Table, StableID string
		Fields                                     []fieldSpec
		SearchIDs                                  []string
	}{
		"billing-campaign-rest-query-catalog.v1", ResourceName, "platform",
		"billing_campaign_query_v1", fieldID, fields, []string{fieldCampaignKey},
	}
	canonical, err := json.Marshal(document)
	if err != nil {
		return api.CatalogHash{}, ErrContract
	}
	digest := sha256.Sum256(canonical)
	hash, err := api.NewCatalogHash(hex.EncodeToString(digest[:]))
	if err != nil {
		return api.CatalogHash{}, ErrContract
	}
	return hash, nil
}

func (catalog *Catalog) Validate() error {
	if catalog == nil || !catalog.valid || catalog.reference.Validate() != nil ||
		catalog.set.Validate() != nil {
		return ErrContract
	}
	return nil
}

func (catalog *Catalog) Catalog(
	ctx context.Context, scope transaction.GlobalScope, resource string,
) (api.CatalogSet, error) {
	if catalog.Validate() != nil || ctx == nil || ctx.Err() != nil ||
		resource != ResourceName || transaction.NetworkAllowed(ctx) ||
		!transaction.ContextOwnsGlobalScope(ctx, scope) ||
		transaction.RequireGlobalOwner(scope, Owner) != nil {
		return api.CatalogSet{}, ErrContract
	}
	return catalog.set, nil
}

func (catalog *Catalog) Reference() (api.CatalogReference, error) {
	if catalog.Validate() != nil {
		return api.CatalogReference{}, ErrContract
	}
	return catalog.reference, nil
}

func (catalog *Catalog) Matches(reference api.CatalogReference) bool {
	if catalog.Validate() != nil || reference.Validate() != nil {
		return false
	}
	leftHash, leftErr := catalog.reference.CatalogHash()
	rightHash, rightErr := reference.CatalogHash()
	leftHex, leftHexErr := leftHash.Hex()
	rightHex, rightHexErr := rightHash.Hex()
	leftOwner, leftOwnerErr := catalog.reference.Owner()
	rightOwner, rightOwnerErr := reference.Owner()
	leftResource, leftResourceErr := catalog.reference.ResourceID()
	rightResource, rightResourceErr := reference.ResourceID()
	leftVersion, leftVersionErr := catalog.reference.PublishedVersionID()
	rightVersion, rightVersionErr := reference.PublishedVersionID()
	return leftErr == nil && rightErr == nil && leftHexErr == nil && rightHexErr == nil &&
		leftOwnerErr == nil && rightOwnerErr == nil && leftResourceErr == nil &&
		rightResourceErr == nil && leftVersionErr == nil && rightVersionErr == nil &&
		leftOwner == rightOwner && leftResource == rightResource &&
		leftVersion == rightVersion && leftHex == rightHex
}
