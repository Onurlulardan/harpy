package httpquery

import restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"

const (
	maximumWireBooleanChildren  = 50
	maximumWireFilterDepth      = 5
	maximumWireFilterNodes      = 100
	maximumWireFilterPredicates = 50
	maximumWireFilterLiterals   = 1000
	maximumWireOperandCount     = 100
)

type filterMetrics struct {
	depth      int
	nodes      int
	predicates int
	literals   int
}

func predicateMetrics(literals int) filterMetrics {
	return filterMetrics{nodes: 1, predicates: 1, literals: literals}
}

func booleanMetrics(children []decodedFilter) filterMetrics {
	result := filterMetrics{depth: 1, nodes: 1}
	for _, child := range children {
		result.depth = max(result.depth, child.metrics.depth+1)
		result.nodes += child.metrics.nodes
		result.predicates += child.metrics.predicates
		result.literals += child.metrics.literals
	}
	return result
}

func validateFilterMetrics(
	metrics filterMetrics, path string, collector *issueCollector,
) bool {
	valid := true
	checks := []struct {
		exceeded bool
		code     restqueryv1.IssueCode
	}{
		{metrics.depth > maximumWireFilterDepth, restqueryv1.IssueBooleanDepthExceeded},
		{metrics.nodes > maximumWireFilterNodes, restqueryv1.IssueASTNodeCountExceeded},
		{metrics.predicates > maximumWireFilterPredicates, restqueryv1.IssuePredicateCountExceeded},
		{metrics.literals > maximumWireFilterLiterals, restqueryv1.IssueLiteralCountExceeded},
	}
	for _, check := range checks {
		if check.exceeded {
			collector.addBudget(path, check.code)
			valid = false
		}
	}
	return valid
}
