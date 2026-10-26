package resolved

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

type filterMetrics struct {
	depth      int
	nodes      int
	predicates int
	literals   int
	valid      bool
}

func measureFilter(filter Filter, booleanDepth int) filterMetrics {
	if booleanDepth > 5 || !filter.validated {
		return filterMetrics{}
	}
	if filter.kind == api.RequestFilterPredicate {
		if filter.predicate == nil || filter.children != nil ||
			filter.predicate.Validate() != nil || len(filter.predicate.values) > 100 {
			return filterMetrics{}
		}
		return filterMetrics{
			nodes: 1, predicates: 1, literals: len(filter.predicate.values), valid: true,
		}
	}
	if filter.predicate != nil || len(filter.children) < 1 || len(filter.children) > 50 ||
		filter.kind != api.RequestFilterAnd && filter.kind != api.RequestFilterOr &&
			filter.kind != api.RequestFilterNot ||
		filter.kind == api.RequestFilterNot && len(filter.children) != 1 {
		return filterMetrics{}
	}
	result := filterMetrics{depth: 1, nodes: 1, valid: true}
	for _, child := range filter.children {
		childMetrics := measureFilter(child, booleanDepth+1)
		if !childMetrics.valid {
			return filterMetrics{}
		}
		result.depth = max(result.depth, childMetrics.depth+1)
		result.nodes += childMetrics.nodes
		result.predicates += childMetrics.predicates
		result.literals += childMetrics.literals
	}
	if result.depth > 5 || result.nodes > 100 || result.predicates > 50 ||
		result.literals > 1000 {
		return filterMetrics{}
	}
	return result
}

func filterCountSafe(filter Filter) bool {
	if filter.predicate != nil {
		capabilities, err := filter.predicate.field.Capabilities()
		if err != nil {
			return false
		}
		capability, found, err := capabilities.Filter(filter.predicate.operator)
		if err != nil || !found {
			return false
		}
		countSafe, err := capability.CountSafe()
		return err == nil && countSafe
	}
	for _, child := range filter.children {
		if !filterCountSafe(child) {
			return false
		}
	}
	return true
}
