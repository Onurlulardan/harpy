package resolved

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func recordPathsCanonical(record RecordQuery, stableID api.FieldDefinition) bool {
	if record.countPath.value != "/count" ||
		!projectionPathsCanonical(record.projections, stableID, "/select") ||
		!sortPathsCanonical(record.sorts, "/sort") ||
		!filterPathCanonical(record.where, "/where") {
		return false
	}
	for _, term := range record.search {
		if term.path.value != "/search" {
			return false
		}
	}
	for index, include := range record.includes {
		if !includePathsCanonical(include, fmt.Sprintf("/include/%d", index)) {
			return false
		}
	}
	return true
}

func includePathsCanonical(include Include, base string) bool {
	if include.path.value != base ||
		!projectionPathsCanonical(include.projections, include.stableID, base+"/select") ||
		!filterPathCanonical(include.where, base+"/where") ||
		!sortPathsCanonical(include.sorts, base+"/sort") {
		return false
	}
	for index, child := range include.children {
		childBase := fmt.Sprintf("%s/include/%d", base, index)
		if !includePathsCanonical(child, childBase) {
			return false
		}
	}
	return true
}

func filterPathCanonical(filter *Filter, expected string) bool {
	if filter == nil {
		return true
	}
	if filter.predicate != nil {
		return filter.predicate.path.value == expected
	}
	for index := range filter.children {
		childExpected := filterChildPath(expected, filter.kind, index)
		if !filterPathCanonical(&filter.children[index], childExpected) {
			return false
		}
	}
	return true
}

func filterChildPath(base string, kind api.RequestFilterKind, index int) string {
	switch kind {
	case api.RequestFilterAnd:
		return fmt.Sprintf("%s/and/%d", base, index)
	case api.RequestFilterOr:
		return fmt.Sprintf("%s/or/%d", base, index)
	case api.RequestFilterNot:
		return base + "/not"
	default:
		return ""
	}
}
