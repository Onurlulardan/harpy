package resolved

import (
	"bytes"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func visibilitiesMatchQuery(
	spec QuerySpec,
	visibilities []ResourceVisibility,
) bool {
	resources, valid := queryAuthorizedResources(spec)
	if !valid || len(resources) != len(visibilities) {
		return false
	}
	for index, visibility := range visibilities {
		if visibility.Validate() != nil {
			return false
		}
		if index > 0 && bytes.Compare(
			visibilities[index-1].requirement[:], visibility.requirement[:],
		) >= 0 {
			return false
		}
		path, err := visibility.path.Value()
		if err != nil {
			return false
		}
		key := authorizedResourceKey{catalog: visibility.catalog, path: path}
		catalog, found := resources[key]
		if !found || !mandatoryFilterMatchesCatalog(visibility.filter, catalog) {
			return false
		}
		delete(resources, key)
	}
	return len(resources) == 0
}

func mandatoryFilterMatchesCatalog(
	filter *MandatoryFilter,
	catalog api.FieldCatalog,
) bool {
	if filter == nil {
		return true
	}
	if filter.Validate() != nil {
		return false
	}
	if filter.predicate != nil &&
		!catalogHasExactField(catalog, filter.predicate.field) {
		return false
	}
	for index := range filter.children {
		if !mandatoryFilterMatchesCatalog(&filter.children[index], catalog) {
			return false
		}
	}
	return true
}
