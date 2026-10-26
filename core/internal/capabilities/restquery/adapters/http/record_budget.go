package httpquery

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type requestFilterTotals struct {
	nodes      int
	predicates int
	literals   int
}

func validateRecordCombinedBudgets(
	selects []string,
	where *api.RequestFilter,
	includes []api.RequestInclude,
	count api.CountMode,
	collector *issueCollector,
) bool {
	valid := true
	if len(selects)+countIncludeFields(includes) > 50 {
		collector.addBudget("/include", restqueryv1.IssueSelectCountExceeded)
		valid = false
	}
	totals := requestFilterTotals{}
	if where != nil {
		totals.add(*where)
	}
	for _, include := range includes {
		totals.addInclude(include)
	}
	checks := []struct {
		exceeded bool
		code     restqueryv1.IssueCode
	}{
		{totals.nodes > 100, restqueryv1.IssueASTNodeCountExceeded},
		{totals.predicates > 50, restqueryv1.IssuePredicateCountExceeded},
		{totals.literals > 1000, restqueryv1.IssueLiteralCountExceeded},
	}
	for _, check := range checks {
		if check.exceeded {
			collector.addBudget("", check.code)
			valid = false
		}
	}
	if count == api.CountModeExact && where != nil {
		valid = validateExactCountFilter(*where, collector) && valid
	}
	return valid
}

func validateExactCountFilter(filter api.RequestFilter, collector *issueCollector) bool {
	valid := true
	totals := requestFilterTotals{}
	totals.add(filter)
	if filter.Depth() > 3 {
		collector.addCountBudget("/where", restqueryv1.IssueBooleanDepthExceeded)
		valid = false
	}
	if totals.predicates > 20 {
		collector.addCountBudget("/where", restqueryv1.IssuePredicateCountExceeded)
		valid = false
	}
	return valid
}

func (totals *requestFilterTotals) add(filter api.RequestFilter) {
	totals.nodes++
	if predicate, found := filter.Predicate(); found {
		totals.predicates++
		totals.literals += len(predicate.Values())
		return
	}
	for _, child := range filter.Children() {
		totals.add(child)
	}
}

func (totals *requestFilterTotals) addInclude(include api.RequestInclude) {
	if where, found := include.Where(); found {
		totals.add(where)
	}
	for _, child := range include.Includes() {
		totals.addInclude(child)
	}
}

func countIncludeFields(includes []api.RequestInclude) int {
	count := 0
	for _, include := range includes {
		count += len(include.Select()) + countIncludeFields(include.Includes())
	}
	return count
}
