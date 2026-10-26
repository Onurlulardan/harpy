package httpquery

import (
	"strconv"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func (document *getDocument) addGETFilter(
	leaf getLeaf, collector *issueCollector,
) {
	if len(leaf.segments) < 2 {
		collector.add("/filter", restqueryv1.IssueIndexInvalid)
		return
	}
	indexPath := appendJSONPointer("/filter", leaf.segments[1])
	index, valid := decodeGETIndex(leaf.segments[1], indexPath, collector)
	if !valid {
		return
	}
	if len(leaf.segments) < 3 {
		collector.add(indexPath, restqueryv1.IssueValueNotCanonical)
		return
	}
	item := document.filters[index]
	if item == nil {
		item = &getFilterDocument{}
		document.filters[index] = item
	}
	switch leaf.segments[2] {
	case "field", "op", "value":
		if len(leaf.segments) != 3 {
			collector.add(leaf.pointer, restqueryv1.IssueValueNotCanonical)
			return
		}
		member := memberFromGETLeaf(leaf)
		if leaf.segments[2] == "field" {
			item.field = member
		} else if leaf.segments[2] == "op" {
			item.operator = member
		} else {
			item.value = member
		}
	case "values":
		document.addGETFilterValue(item, leaf, collector)
	case "and", "or", "not":
		collector.add(leaf.pointer, restqueryv1.IssueMemberForbidden)
	default:
		collector.add(leaf.pointer, restqueryv1.IssueMemberUnknown)
	}
}

func (document *getDocument) addGETFilterValue(
	item *getFilterDocument, leaf getLeaf, collector *issueCollector,
) {
	valuesPath := getFilterValuesPath(leaf.segments[1])
	if len(leaf.segments) < 4 {
		collector.add(valuesPath, restqueryv1.IssueIndexInvalid)
		return
	}
	index, valid := decodeGETIndex(
		leaf.segments[3], appendJSONPointer(valuesPath, leaf.segments[3]), collector,
	)
	if !valid {
		return
	}
	if len(leaf.segments) != 4 {
		collector.add(leaf.pointer, restqueryv1.IssueValueNotCanonical)
		return
	}
	if item.values == nil {
		item.values = make(map[int]getMember)
	}
	item.values[index] = *memberFromGETLeaf(leaf)
}

func decodeGETFilters(
	document getDocument, collector *issueCollector,
) (*api.RequestFilter, bool) {
	indices, valid := orderedGETIndices(document.filters, "/filter", collector)
	if len(indices) > 50 {
		collector.addBudget("/filter", restqueryv1.IssuePredicateCountExceeded)
		valid = false
	}
	predicates := make([]api.RequestFilter, 0, len(indices))
	for _, index := range indices {
		predicate, predicateValid := decodeGETFilter(
			document.filters[index], "/filter/"+strconv.Itoa(index), collector,
		)
		if !predicateValid {
			valid = false
			continue
		}
		predicates = append(predicates, predicate)
	}
	if len(predicates) == 0 {
		return nil, valid
	}
	where, err := api.NewRequestAnd(predicates)
	if err != nil {
		collector.addBudget("/filter", restqueryv1.IssuePredicateCountExceeded)
		return nil, false
	}
	return &where, valid
}

func decodeGETFilter(
	item *getFilterDocument, path string, collector *issueCollector,
) (api.RequestFilter, bool) {
	field, fieldValid := requiredGETFilterField(item, path, collector)
	operator, operatorValid := requiredGETOperator(item, path, collector)
	if !operatorValid {
		return api.RequestFilter{}, false
	}
	values, valuesValid := decodeGETOperands(item, path, operator, collector)
	if !fieldValid || !valuesValid {
		return api.RequestFilter{}, false
	}
	predicate, err := api.NewRequestPredicate(field, operator, values)
	if err != nil {
		collector.add(path, restqueryv1.IssueValueNotAllowed)
		return api.RequestFilter{}, false
	}
	return predicate, true
}

func getFilterValuesPath(index string) string {
	return "/filter/" + index + "/values"
}
