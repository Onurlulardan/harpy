package resolved

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func aggregateMembersMatchCatalog(
	query AggregateQuery, catalog api.FieldCatalog,
) bool {
	for _, group := range query.groups {
		if !catalogHasExactField(catalog, group.field) {
			return false
		}
		publicName, err := group.field.PublicName()
		key, keyErr := group.key.Value()
		if err != nil || keyErr != nil || publicName != key {
			return false
		}
	}
	for _, measure := range query.measures {
		if measure.hasField && !catalogHasExactField(catalog, measure.field) {
			return false
		}
	}
	return filterMatchesCatalog(query.where, catalog) &&
		searchMatchesCatalog(query.search, catalog)
}
