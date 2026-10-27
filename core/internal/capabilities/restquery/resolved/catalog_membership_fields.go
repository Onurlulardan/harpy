package resolved

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func catalogHasExactField(catalog api.FieldCatalog, field api.FieldDefinition) bool {
	identifier, err := field.ID()
	if err != nil {
		return false
	}
	declared, found, err := catalog.Field(identifier)
	return err == nil && found && sameFieldDefinition(declared, field)
}

func projectionsMatchCatalog(projections []Projection, catalog api.FieldCatalog) bool {
	for _, projection := range projections {
		if !catalogHasExactField(catalog, projection.field) {
			return false
		}
		publicName, err := projection.field.PublicName()
		outputKey, keyErr := projection.key.Value()
		if err != nil || keyErr != nil || publicName != outputKey {
			return false
		}
	}
	return true
}

func filterMatchesCatalog(filter *Filter, catalog api.FieldCatalog) bool {
	if filter == nil {
		return true
	}
	if filter.predicate != nil && !catalogHasExactField(catalog, filter.predicate.field) {
		return false
	}
	for index := range filter.children {
		if !filterMatchesCatalog(&filter.children[index], catalog) {
			return false
		}
	}
	return true
}

func sortsMatchCatalog(sorts []Sort, catalog api.FieldCatalog) bool {
	for _, sort := range sorts {
		if !catalogHasExactField(catalog, sort.field) {
			return false
		}
	}
	return true
}

func searchMatchesCatalog(terms []SearchTerm, catalog api.FieldCatalog) bool {
	for _, term := range terms {
		if !catalogHasExactField(catalog, term.field) {
			return false
		}
	}
	return true
}

func sameRelationDefinition(left, right api.RelationDefinition) bool {
	leftID, leftIDErr := left.RelationID()
	rightID, rightIDErr := right.RelationID()
	leftName, leftNameErr := left.PublicName()
	rightName, rightNameErr := right.PublicName()
	leftCardinality, leftCardinalityErr := left.Cardinality()
	rightCardinality, rightCardinalityErr := right.Cardinality()
	leftStorage, leftStorageErr := left.Storage()
	rightStorage, rightStorageErr := right.Storage()
	leftTarget, leftTargetErr := left.Target()
	rightTarget, rightTargetErr := right.Target()
	leftEnabled, leftEnabledErr := left.IncludeEnabled()
	rightEnabled, rightEnabledErr := right.IncludeEnabled()
	leftMaximum, leftMaximumErr := left.MaximumIncludedRecords()
	rightMaximum, rightMaximumErr := right.MaximumIncludedRecords()
	return leftIDErr == nil && rightIDErr == nil && leftNameErr == nil &&
		rightNameErr == nil && leftCardinalityErr == nil && rightCardinalityErr == nil &&
		leftStorageErr == nil && rightStorageErr == nil &&
		leftTargetErr == nil && rightTargetErr == nil && leftEnabledErr == nil &&
		rightEnabledErr == nil && leftMaximumErr == nil && rightMaximumErr == nil &&
		leftID == rightID && leftName == rightName &&
		leftCardinality == rightCardinality && leftStorage == rightStorage &&
		sameCatalogReference(leftTarget, rightTarget) &&
		leftEnabled == rightEnabled && leftMaximum == rightMaximum
}
