package httpquery

import (
	"encoding/json"
	"strconv"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func decodeHavingBoolean(
	name string,
	raw json.RawMessage,
	path string,
	groups map[string]struct{},
	measures map[string]struct{},
	collector *issueCollector,
) (decodedHaving, bool) {
	memberPath := rawMemberPath(path, name)
	if name == "not" {
		child, valid := decodeHaving(raw, memberPath, groups, measures, collector)
		metrics := booleanHavingMetrics([]decodedHaving{child})
		if !valid || !validateHavingMetrics(metrics, path, collector) {
			return decodedHaving{metrics: metrics}, false
		}
		having, err := api.NewRequestHavingNot(child.having)
		if err != nil {
			collector.add(path, restqueryv1.IssueValueNotAllowed)
			return decodedHaving{metrics: metrics}, false
		}
		return decodedHaving{having: having, metrics: metrics}, true
	}
	items, arrayValid := decodeArray(raw, memberPath, collector)
	if !arrayValid {
		return decodedHaving{}, false
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
	children := make([]decodedHaving, 0, len(items))
	for index, item := range items {
		itemPath := rawMemberPath(memberPath, strconv.Itoa(index))
		child, childValid := decodeHaving(item, itemPath, groups, measures, collector)
		children = append(children, child)
		valid = valid && childValid
	}
	metrics := booleanHavingMetrics(children)
	valid = validateHavingMetrics(metrics, path, collector) && valid
	if !valid {
		return decodedHaving{metrics: metrics}, false
	}
	values := make([]api.RequestHaving, len(children))
	for index := range children {
		values[index] = children[index].having
	}
	having, err := newHavingBoolean(name, values)
	if err != nil {
		collector.add(path, restqueryv1.IssueValueNotAllowed)
		return decodedHaving{metrics: metrics}, false
	}
	return decodedHaving{having: having, metrics: metrics}, true
}

func newHavingBoolean(name string, children []api.RequestHaving) (api.RequestHaving, error) {
	if name == "and" {
		return api.NewRequestHavingAnd(children)
	}
	return api.NewRequestHavingOr(children)
}
