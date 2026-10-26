package catalog

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

type Published struct {
	resource        Resource
	reference       api.CatalogReference
	set             api.CatalogSet
	classifications map[uuid.UUID]profile.Classification
	createRequired  map[uuid.UUID]bool
	maximumBytes    map[uuid.UUID]int
	validated       bool
}

func (published Published) Validate() error {
	if !published.validated || published.set.Validate() != nil ||
		published.reference.Validate() != nil || len(published.classifications) == 0 ||
		len(published.createRequired) != len(published.classifications) ||
		len(published.maximumBytes) != len(published.classifications) {
		return ErrContract
	}
	root, err := published.set.Root()
	if err != nil || !sameReference(root, published.reference) {
		return ErrContract
	}
	resourceID, err := root.ResourceID()
	if err != nil || resourceID != published.resource.ID() {
		return ErrContract
	}
	catalog, err := published.set.RootCatalog()
	if err != nil {
		return ErrContract
	}
	fields, err := catalog.Fields()
	if err != nil || len(fields) != len(published.classifications) {
		return ErrContract
	}
	for _, field := range fields {
		identifier, fieldErr := field.ID()
		classification, found := published.classifications[identifier]
		_, requiredFound := published.createRequired[identifier]
		maximumBytes, maximumFound := published.maximumBytes[identifier]
		if fieldErr != nil || !found || !classification.Valid() || !requiredFound ||
			!maximumFound || maximumBytes < 0 ||
			maximumBytes > profile.MaximumDocumentBytes {
			return ErrContract
		}
	}
	return nil
}

func (published Published) CatalogSet() (api.CatalogSet, error) {
	if published.Validate() != nil {
		return api.CatalogSet{}, ErrContract
	}
	return published.set, nil
}

func (published Published) Reference() (api.CatalogReference, error) {
	if published.Validate() != nil {
		return api.CatalogReference{}, ErrContract
	}
	return published.reference, nil
}

func (published Published) Resource() (Resource, error) {
	if published.Validate() != nil {
		return Resource{}, ErrContract
	}
	return published.resource, nil
}

func (published Published) Matches(reference api.CatalogReference) bool {
	return published.Validate() == nil && reference.Validate() == nil &&
		sameReference(published.reference, reference)
}

func (published Published) Classification(
	fieldID uuid.UUID,
) (profile.Classification, error) {
	if published.Validate() != nil {
		return "", ErrContract
	}
	classification, found := published.classifications[fieldID]
	if !found || !classification.Valid() {
		return "", ErrContract
	}
	return classification, nil
}

func (published Published) RequiredForCreate(fieldID uuid.UUID) (bool, error) {
	if published.Validate() != nil {
		return false, ErrContract
	}
	required, found := published.createRequired[fieldID]
	if !found {
		return false, ErrContract
	}
	return required, nil
}

func (published Published) MaximumInputBytes(fieldID uuid.UUID) (int, error) {
	if published.Validate() != nil {
		return 0, ErrContract
	}
	maximum, found := published.maximumBytes[fieldID]
	if !found {
		return 0, ErrContract
	}
	return maximum, nil
}

func sameReference(left, right api.CatalogReference) bool {
	leftID, leftErr := left.ResourceID()
	rightID, rightErr := right.ResourceID()
	leftVersion, leftVersionErr := left.PublishedVersionID()
	rightVersion, rightVersionErr := right.PublishedVersionID()
	leftHash, leftHashErr := left.CatalogHash()
	rightHash, rightHashErr := right.CatalogHash()
	leftHex, leftHexErr := leftHash.Hex()
	rightHex, rightHexErr := rightHash.Hex()
	leftOwner, leftOwnerErr := left.Owner()
	rightOwner, rightOwnerErr := right.Owner()
	return leftErr == nil && rightErr == nil && leftVersionErr == nil &&
		rightVersionErr == nil && leftHashErr == nil && rightHashErr == nil &&
		leftHexErr == nil && rightHexErr == nil && leftOwnerErr == nil &&
		rightOwnerErr == nil && leftID == rightID && leftVersion == rightVersion &&
		leftHex == rightHex && leftOwner == rightOwner
}
