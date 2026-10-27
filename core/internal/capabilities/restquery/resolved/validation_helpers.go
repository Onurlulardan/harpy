package resolved

import (
	"bytes"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func sameCatalogReference(left, right api.CatalogReference) bool {
	leftOwner, leftErr := left.Owner()
	rightOwner, rightErr := right.Owner()
	leftResource, leftResourceErr := left.ResourceID()
	rightResource, rightResourceErr := right.ResourceID()
	leftVersion, leftVersionErr := left.PublishedVersionID()
	rightVersion, rightVersionErr := right.PublishedVersionID()
	leftHash, leftHashErr := left.CatalogHash()
	rightHash, rightHashErr := right.CatalogHash()
	leftBytes, leftBytesErr := leftHash.Bytes()
	rightBytes, rightBytesErr := rightHash.Bytes()
	return leftErr == nil && rightErr == nil && leftResourceErr == nil &&
		rightResourceErr == nil && leftVersionErr == nil && rightVersionErr == nil &&
		leftHashErr == nil && rightHashErr == nil && leftBytesErr == nil &&
		rightBytesErr == nil && leftOwner == rightOwner && leftResource == rightResource &&
		leftVersion == rightVersion && bytes.Equal(leftBytes, rightBytes)
}

func fieldIdentifier(field api.FieldDefinition) (uuid.UUID, bool) {
	identifier, err := field.ID()
	return identifier, err == nil
}

func validProjectionSet(projections []Projection, allowHidden bool) bool {
	fieldIDs := make(map[uuid.UUID]struct{}, len(projections))
	keys := make(map[string]struct{}, len(projections))
	for _, projection := range projections {
		if projection.Validate() != nil || !allowHidden && projection.hidden {
			return false
		}
		fieldID, valid := fieldIdentifier(projection.field)
		key, keyErr := projection.key.Value()
		if !valid || keyErr != nil {
			return false
		}
		if _, found := fieldIDs[fieldID]; found {
			return false
		}
		if _, found := keys[key]; found {
			return false
		}
		fieldIDs[fieldID], keys[key] = struct{}{}, struct{}{}
	}
	return true
}

func containsField(projections []Projection, wanted api.FieldDefinition) bool {
	wantedID, valid := fieldIdentifier(wanted)
	if !valid {
		return false
	}
	for _, projection := range projections {
		fieldID, fieldValid := fieldIdentifier(projection.field)
		if fieldValid && fieldID == wantedID {
			return true
		}
	}
	return false
}
