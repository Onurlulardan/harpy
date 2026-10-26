package v1query

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type filterBudget struct {
	nodes      int
	predicates int
	literals   int
}

func measureFilters(children []api.RequestFilter) filterBudget {
	result := filterBudget{nodes: 1}
	for _, child := range children {
		measured := measureFilter(child)
		result.nodes += measured.nodes
		result.predicates += measured.predicates
		result.literals += measured.literals
	}
	return result
}

func measureFilter(filter api.RequestFilter) filterBudget {
	if predicate, found := filter.Predicate(); found {
		return filterBudget{nodes: 1, predicates: 1, literals: len(predicate.Values())}
	}
	result := filterBudget{nodes: 1}
	for _, child := range filter.Children() {
		measured := measureFilter(child)
		result.nodes += measured.nodes
		result.predicates += measured.predicates
		result.literals += measured.literals
	}
	return result
}

func addFilterBudgetIssues(
	children []api.RequestFilter, collector *issueCollector,
) bool {
	measured := measureFilters(children)
	exceeded := false
	if measured.predicates > 50 {
		collector.addBudget("/filters", restqueryv1.IssuePredicateCountExceeded)
		exceeded = true
	}
	if measured.literals > 1000 {
		collector.addBudget("/filters", restqueryv1.IssueLiteralCountExceeded)
		exceeded = true
	}
	if measured.nodes > 100 {
		collector.addBudget("/filters", restqueryv1.IssueASTNodeCountExceeded)
		exceeded = true
	}
	return exceeded
}
