package reports

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func (catalog *Catalog) resourceFor(
	reference api.CatalogReference,
) (resourceSpec, bool) {
	if catalog.Validate() != nil || reference.Validate() != nil {
		return resourceSpec{}, false
	}
	for _, entry := range catalog.entries {
		if sameReference(entry.reference, reference) {
			return entry.spec, true
		}
	}
	return resourceSpec{}, false
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
	return leftOwnerErr == nil && rightOwnerErr == nil && leftResourceErr == nil &&
		rightResourceErr == nil && leftVersionErr == nil && rightVersionErr == nil &&
		leftHashErr == nil && rightHashErr == nil && leftHexErr == nil &&
		rightHexErr == nil && leftOwner == rightOwner && leftResource == rightResource &&
		leftVersion == rightVersion && leftHex == rightHex
}
