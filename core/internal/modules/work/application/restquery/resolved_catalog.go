package restquery

import (
	restapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

type ResolvedCatalog struct {
	resource  string
	reference restapi.CatalogReference
	set       restapi.CatalogSet
	fields    map[string]uuid.UUID
}

func NewResolvedCatalog(resource string, set restapi.CatalogSet) (ResolvedCatalog, error) {
	if operationFor(resource) == "" || set.Validate() != nil {
		return ResolvedCatalog{}, ErrInvalid
	}
	reference, err := set.Root()
	owner, ownerErr := reference.Owner()
	root, rootErr := set.RootCatalog()
	if err != nil || ownerErr != nil || rootErr != nil || owner != Owner {
		return ResolvedCatalog{}, ErrInvalid
	}
	definitions, err := root.Fields()
	if err != nil || len(definitions) == 0 {
		return ResolvedCatalog{}, ErrInvalid
	}
	fields := make(map[string]uuid.UUID, len(definitions))
	for _, definition := range definitions {
		name, nameErr := definition.PublicName()
		identifier, identifierErr := definition.ID()
		if nameErr != nil || identifierErr != nil || identifier == uuid.Nil {
			return ResolvedCatalog{}, ErrInvalid
		}
		fields[name] = identifier
	}
	return ResolvedCatalog{resource: resource, reference: reference, set: set,
		fields: fields}, nil
}

func (catalog ResolvedCatalog) Validate() error {
	if operationFor(catalog.resource) == "" || catalog.reference.Validate() != nil ||
		catalog.set.Validate() != nil || len(catalog.fields) == 0 {
		return ErrInvalid
	}
	root, err := catalog.set.Root()
	if err != nil || !sameCatalogReference(root, catalog.reference) {
		return ErrInvalid
	}
	return nil
}

func (catalog ResolvedCatalog) fieldID(name string) (uuid.UUID, error) {
	if catalog.Validate() != nil {
		return uuid.Nil, ErrUnavailable
	}
	identifier := catalog.fields[name]
	if identifier == uuid.Nil {
		return uuid.Nil, ErrUnavailable
	}
	return identifier, nil
}

func (catalog ResolvedCatalog) matches(reference restapi.CatalogReference) bool {
	return catalog.Validate() == nil && sameCatalogReference(reference, catalog.reference)
}

func sameCatalogReference(left, right restapi.CatalogReference) bool {
	leftResource, leftResourceErr := left.ResourceID()
	rightResource, rightResourceErr := right.ResourceID()
	leftVersion, leftVersionErr := left.PublishedVersionID()
	rightVersion, rightVersionErr := right.PublishedVersionID()
	leftHash, leftHashErr := left.CatalogHash()
	rightHash, rightHashErr := right.CatalogHash()
	leftHex, leftHexErr := leftHash.Hex()
	rightHex, rightHexErr := rightHash.Hex()
	return leftResourceErr == nil && rightResourceErr == nil &&
		leftVersionErr == nil && rightVersionErr == nil && leftHashErr == nil &&
		rightHashErr == nil && leftHexErr == nil && rightHexErr == nil &&
		leftResource == rightResource && leftVersion == rightVersion && leftHex == rightHex
}
