package resolved

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func aggregatePathsCanonical(query AggregateQuery) bool {
	for index, group := range query.groups {
		if group.path.value != fmt.Sprintf("/aggregate/group_by/%d", index) {
			return false
		}
	}
	for index, measure := range query.measures {
		if measure.path.value != fmt.Sprintf("/aggregate/measures/%d", index) {
			return false
		}
	}
	if !filterPathCanonical(query.where, "/where") ||
		!havingPathCanonical(query.having, "/aggregate/having") {
		return false
	}
	for _, term := range query.search {
		if term.path.value != "/search" {
			return false
		}
	}
	return aggregateSortPathsCanonical(query.sorts)
}

func aggregateSortPathsCanonical(sorts []AggregateSort) bool {
	explicitIndex := 0
	for _, sort := range sorts {
		path := sort.reference.path.value
		if sort.implicit {
			if path != "/aggregate/sort" {
				return false
			}
			continue
		}
		if path != fmt.Sprintf("/aggregate/sort/%d", explicitIndex) {
			return false
		}
		explicitIndex++
	}
	return true
}

func havingPathCanonical(having *Having, expected string) bool {
	if having == nil {
		return true
	}
	if having.predicate != nil {
		return having.predicate.reference.path.value == expected
	}
	for index := range having.children {
		childExpected := havingChildPath(expected, having.kind, index)
		if !havingPathCanonical(&having.children[index], childExpected) {
			return false
		}
	}
	return true
}

func havingChildPath(base string, kind api.RequestHavingKind, index int) string {
	switch kind {
	case api.RequestHavingKindAnd:
		return fmt.Sprintf("%s/and/%d", base, index)
	case api.RequestHavingKindOr:
		return fmt.Sprintf("%s/or/%d", base, index)
	case api.RequestHavingKindNot:
		return base + "/not"
	default:
		return ""
	}
}
