package httpquery

import restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"

type havingMetrics struct {
	depth      int
	nodes      int
	predicates int
	literals   int
}

func booleanHavingMetrics(children []decodedHaving) havingMetrics {
	result := havingMetrics{depth: 1, nodes: 1}
	for _, child := range children {
		result.depth = max(result.depth, child.metrics.depth+1)
		result.nodes += child.metrics.nodes
		result.predicates += child.metrics.predicates
		result.literals += child.metrics.literals
	}
	return result
}

func validateHavingMetrics(
	metrics havingMetrics, path string, collector *issueCollector,
) bool {
	valid := true
	if metrics.depth > 3 {
		collector.addBudget(path, restqueryv1.IssueBooleanDepthExceeded)
		valid = false
	}
	if metrics.predicates > 10 {
		collector.addBudget(path, restqueryv1.IssueHavingCountExceeded)
		valid = false
	}
	if metrics.nodes > 100 {
		collector.addBudget(path, restqueryv1.IssueASTNodeCountExceeded)
		valid = false
	}
	if metrics.literals > 1000 {
		collector.addBudget(path, restqueryv1.IssueLiteralCountExceeded)
		valid = false
	}
	return valid
}
