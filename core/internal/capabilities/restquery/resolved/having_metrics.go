package resolved

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

type havingMetrics struct {
	depth      int
	nodes      int
	predicates int
	literals   int
	valid      bool
}

func measureHaving(having Having, booleanDepth int) havingMetrics {
	if booleanDepth > 3 || !having.validated {
		return havingMetrics{}
	}
	if having.kind == api.RequestHavingKindPredicate {
		if having.predicate == nil || having.children != nil ||
			having.predicate.Validate() != nil || len(having.predicate.values) > 100 {
			return havingMetrics{}
		}
		return havingMetrics{
			nodes: 1, predicates: 1, literals: len(having.predicate.values), valid: true,
		}
	}
	validKind := having.kind == api.RequestHavingKindAnd ||
		having.kind == api.RequestHavingKindOr || having.kind == api.RequestHavingKindNot
	if !validKind || having.predicate != nil || len(having.children) < 1 ||
		len(having.children) > 50 ||
		having.kind == api.RequestHavingKindNot && len(having.children) != 1 {
		return havingMetrics{}
	}
	result := havingMetrics{depth: 1, nodes: 1, valid: true}
	for _, child := range having.children {
		childMetrics := measureHaving(child, booleanDepth+1)
		if !childMetrics.valid {
			return havingMetrics{}
		}
		result.depth = max(result.depth, childMetrics.depth+1)
		result.nodes += childMetrics.nodes
		result.predicates += childMetrics.predicates
		result.literals += childMetrics.literals
	}
	if result.depth > 3 || result.nodes > 100 || result.predicates > 10 ||
		result.literals > 1000 {
		return havingMetrics{}
	}
	return result
}
