package httpquery

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func decodeHavingPredicate(
	object rawObject,
	path string,
	groups map[string]struct{},
	measures map[string]struct{},
	collector *issueCollector,
) (decodedHaving, bool) {
	reference, name, referenceValid := decodeAggregateReference(
		object, path, groups, measures, collector,
	)
	opRaw, operatorPresent := requireMember(object, path, "op", collector)
	operator, operatorValid := decodeOperatorIfPresent(
		opRaw, operatorPresent, rawMemberPath(path, "op"), collector,
	)
	values, valuesValid := decodePredicateOperands(object, path, operator, operatorValid, collector)
	metrics := havingMetrics{nodes: 1, predicates: 1, literals: len(values)}
	metricsValid := validateHavingMetrics(metrics, path, collector)
	if !referenceValid || !operatorValid || !valuesValid || !metricsValid {
		return decodedHaving{metrics: metrics}, false
	}
	having, err := api.NewRequestHavingPredicate(reference, name, operator, values)
	if err != nil {
		collector.add(path, restqueryv1.IssueValueNotAllowed)
		return decodedHaving{metrics: metrics}, false
	}
	return decodedHaving{having: having, metrics: metrics}, true
}
