package resolved

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func specMembersMatchCatalogs(spec QuerySpec) bool {
	root, err := spec.catalogs.RootCatalog()
	if err != nil || !catalogHasExactField(root, spec.stableID) {
		return false
	}
	if spec.record != nil {
		return recordMembersMatchCatalog(*spec.record, root, spec.catalogs)
	}
	return aggregateMembersMatchCatalog(*spec.aggregate, root)
}

func recordMembersMatchCatalog(
	record RecordQuery, root api.FieldCatalog, catalogs api.CatalogSet,
) bool {
	if !projectionsMatchCatalog(record.projections, root) ||
		!filterMatchesCatalog(record.where, root) ||
		!sortsMatchCatalog(record.sorts, root) ||
		!searchMatchesCatalog(record.search, root) ||
		!recordCountCapabilityMatches(record, root) {
		return false
	}
	for _, include := range record.includes {
		if !includeMembersMatchCatalog(include, root, catalogs) {
			return false
		}
	}
	return true
}

func recordCountCapabilityMatches(record RecordQuery, catalog api.FieldCatalog) bool {
	capabilities, err := catalog.Capabilities()
	if err != nil {
		return false
	}
	exactCountEnabled, err := capabilities.ExactCountEnabled()
	return err == nil && exactCountEnabled == record.exactCountEnabled
}

func includeMembersMatchCatalog(
	include Include, source api.FieldCatalog, catalogs api.CatalogSet,
) bool {
	relationID, err := include.relation.RelationID()
	if err != nil {
		return false
	}
	declared, found, err := source.Relation(relationID)
	if err != nil || !found || !sameRelationDefinition(declared, include.relation) {
		return false
	}
	target, found, err := catalogs.Catalog(include.target)
	if err != nil || !found || !catalogNodeMatchesInclude(target, include) {
		return false
	}
	for _, child := range include.children {
		if !includeMembersMatchCatalog(child, target, catalogs) {
			return false
		}
	}
	return true
}

func catalogNodeMatchesInclude(catalog api.FieldCatalog, include Include) bool {
	storage, storageErr := catalog.Storage()
	stableID, stableErr := catalog.StableID()
	includeStableID, includeStableErr := include.stableID.ID()
	return storageErr == nil && stableErr == nil && includeStableErr == nil &&
		storage == include.targetStorage && stableID == includeStableID &&
		catalogHasExactField(catalog, include.stableID) &&
		projectionsMatchCatalog(include.projections, catalog) &&
		filterMatchesCatalog(include.where, catalog) &&
		sortsMatchCatalog(include.sorts, catalog)
}
