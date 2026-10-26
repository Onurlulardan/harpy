package httpquery

import restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"

func validateAggregateCombinedBudget(
	where filterMetrics, having havingMetrics, collector *issueCollector,
) bool {
	valid := true
	if where.nodes+having.nodes > maximumWireFilterNodes {
		collector.addBudget("/aggregate/having", restqueryv1.IssueASTNodeCountExceeded)
		valid = false
	}
	if where.literals+having.literals > maximumWireFilterLiterals {
		collector.addBudget("/aggregate/having", restqueryv1.IssueLiteralCountExceeded)
		valid = false
	}
	return valid
}
