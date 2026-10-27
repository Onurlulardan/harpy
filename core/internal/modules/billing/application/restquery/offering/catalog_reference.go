package offering

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func (catalog *Catalog) Reference() (api.CatalogReference, error) {
	if catalog.Validate() != nil {
		return api.CatalogReference{}, ErrContract
	}
	return catalog.reference, nil
}

func (catalog *Catalog) Matches(reference api.CatalogReference) bool {
	return catalog.Validate() == nil && reference.Validate() == nil &&
		sameReference(catalog.reference, reference)
}

func sameReference(left, right api.CatalogReference) bool {
	leftOwner, leftOwnerErr := left.Owner()
	rightOwner, rightOwnerErr := right.Owner()
	leftResource, leftResourceErr := left.ResourceID()
	rightResource, rightResourceErr := right.ResourceID()
	leftVersion, leftVersionErr := left.PublishedVersionID()
	rightVersion, rightVersionErr := right.PublishedVersionID()
	leftHash, leftHashErr := left.CatalogHash()
	rightHash, rightHashErr := right.CatalogHash()
	leftHex, leftHexErr := leftHash.Hex()
	rightHex, rightHexErr := rightHash.Hex()
	return leftOwnerErr == nil && rightOwnerErr == nil &&
		leftResourceErr == nil && rightResourceErr == nil &&
		leftVersionErr == nil && rightVersionErr == nil &&
		leftHashErr == nil && rightHashErr == nil &&
		leftHexErr == nil && rightHexErr == nil &&
		leftOwner == rightOwner && leftResource == rightResource &&
		leftVersion == rightVersion && leftHex == rightHex
}
