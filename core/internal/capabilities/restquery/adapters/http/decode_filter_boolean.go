package httpquery

import (
	"encoding/json"
	"strconv"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func decodeBooleanFilter(
	name string, raw json.RawMessage, path string, collector *issueCollector,
) (decodedFilter, bool) {
	memberPath := rawMemberPath(path, name)
	if name == "not" {
		child, valid := decodeFilter(raw, memberPath, collector)
		metrics := booleanMetrics([]decodedFilter{child})
		if !valid || !validateFilterMetrics(metrics, path, collector) {
			return decodedFilter{metrics: metrics}, false
		}
		filter, err := api.NewRequestNot(child.filter)
		if err != nil {
			collector.add(path, restqueryv1.IssueValueNotAllowed)
			return decodedFilter{metrics: metrics}, false
		}
		return decodedFilter{filter: filter, metrics: metrics}, true
	}
	items, arrayValid := decodeArray(raw, memberPath, collector)
	if !arrayValid {
		return decodedFilter{}, false
	}
	valid := true
	if len(items) == 0 {
		collector.add(memberPath, restqueryv1.IssueArityMismatch)
		valid = false
	}
	if len(items) > maximumWireBooleanChildren {
		collector.addBudget(memberPath, restqueryv1.IssueItemCountExceeded)
		valid = false
	}
	children := make([]decodedFilter, 0, len(items))
	for index, item := range items {
		childPath := rawMemberPath(memberPath, strconv.Itoa(index))
		child, childValid := decodeFilter(item, childPath, collector)
		children = append(children, child)
		valid = valid && childValid
	}
	metrics := booleanMetrics(children)
	valid = validateFilterMetrics(metrics, path, collector) && valid
	if !valid {
		return decodedFilter{metrics: metrics}, false
	}
	values := make([]api.RequestFilter, len(children))
	for index := range children {
		values[index] = children[index].filter
	}
	filter, err := newBooleanFilter(name, values)
	if err != nil {
		collector.add(path, restqueryv1.IssueValueNotAllowed)
		return decodedFilter{metrics: metrics}, false
	}
	return decodedFilter{filter: filter, metrics: metrics}, true
}

func newBooleanFilter(name string, children []api.RequestFilter) (api.RequestFilter, error) {
	if name == "and" {
		return api.NewRequestAnd(children)
	}
	return api.NewRequestOr(children)
}
