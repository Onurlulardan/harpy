package httpquery

import (
	"strconv"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

var measureMembers = []string{"name", "function", "field"}

func decodeMeasures(
	raw []byte, present bool, path string, collector *issueCollector,
) ([]api.RequestMeasure, bool) {
	if !present {
		return []api.RequestMeasure{}, true
	}
	items, valid := decodeArray(raw, path, collector)
	if !valid {
		return nil, false
	}
	if len(items) > 8 {
		collector.addBudget(path, restqueryv1.IssueMeasureCountExceeded)
		valid = false
	}
	result := make([]api.RequestMeasure, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for index, item := range items {
		itemPath := rawMemberPath(path, strconv.Itoa(index))
		measure, itemValid := decodeMeasure(item, itemPath, collector)
		if !itemValid {
			valid = false
			continue
		}
		if _, duplicate := seen[measure.Name()]; duplicate {
			collector.add(rawMemberPath(itemPath, "name"), restqueryv1.IssueNameDuplicate)
			valid = false
			continue
		}
		seen[measure.Name()] = struct{}{}
		result = append(result, measure)
	}
	return result, valid
}

func decodeMeasure(
	raw []byte, path string, collector *issueCollector,
) (api.RequestMeasure, bool) {
	start := len(collector.issues)
	object, objectValid := decodeObject(raw, path, measureMembers, collector)
	if !objectValid {
		return api.RequestMeasure{}, false
	}
	nameRaw, namePresent := requireMember(object, path, "name", collector)
	functionRaw, functionPresent := requireMember(object, path, "function", collector)
	name, nameValid := decodeMeasureName(nameRaw, namePresent, rawMemberPath(path, "name"), collector)
	function, functionValid := decodeAggregateFunction(
		functionRaw, functionPresent, rawMemberPath(path, "function"), collector,
	)
	field, fieldValid := decodeMeasureField(object, path, function, functionValid, collector)
	if !nameValid || !functionValid || !fieldValid || len(collector.issues) != start {
		return api.RequestMeasure{}, false
	}
	measure, err := api.NewRequestMeasure(name, function, field)
	if err != nil {
		collector.add(path, restqueryv1.IssueValueNotAllowed)
		return api.RequestMeasure{}, false
	}
	return measure, true
}

func decodeMeasureName(
	raw []byte, present bool, path string, collector *issueCollector,
) (string, bool) {
	if !present {
		return "", false
	}
	name, valid := decodeName(raw, path, collector)
	if !valid {
		return "", false
	}
	if name == reservedIncludedName {
		collector.add(path, restqueryv1.IssueNameReserved)
		return "", false
	}
	if !validMeasureWireName(name) {
		collector.add(path, restqueryv1.IssueNameInvalid)
		return "", false
	}
	return name, true
}
