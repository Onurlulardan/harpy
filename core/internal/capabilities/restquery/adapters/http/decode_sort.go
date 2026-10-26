package httpquery

import (
	"encoding/json"
	"strconv"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

var sortMembers = []string{"field", "direction"}

func decodeSorts(
	raw json.RawMessage, present bool, path string, collector *issueCollector,
) ([]api.RequestSort, bool) {
	if !present {
		return []api.RequestSort{}, true
	}
	items, valid := decodeArray(raw, path, collector)
	if !valid {
		return nil, false
	}
	if len(items) > 5 {
		collector.addBudget(path, restqueryv1.IssueSortCountExceeded)
		valid = false
	}
	result := make([]api.RequestSort, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for index, item := range items {
		itemPath := rawMemberPath(path, strconv.Itoa(index))
		sort, itemValid := decodeSort(item, itemPath, collector)
		if !itemValid {
			valid = false
			continue
		}
		if _, duplicate := seen[sort.Field()]; duplicate {
			collector.add(rawMemberPath(itemPath, "field"), restqueryv1.IssueFieldDuplicate)
			valid = false
			continue
		}
		seen[sort.Field()] = struct{}{}
		result = append(result, sort)
	}
	return result, valid
}

func decodeSort(
	raw []byte, path string, collector *issueCollector,
) (api.RequestSort, bool) {
	start := len(collector.issues)
	object, valid := decodeObject(raw, path, sortMembers, collector)
	if !valid {
		return api.RequestSort{}, false
	}
	fieldRaw, fieldPresent := requireMember(object, path, "field", collector)
	directionRaw, directionPresent := requireMember(object, path, "direction", collector)
	field, fieldValid := decodeFieldNameIfPresent(
		fieldRaw, fieldPresent, rawMemberPath(path, "field"), collector,
	)
	direction, directionValid := decodeDirection(
		directionRaw, directionPresent, rawMemberPath(path, "direction"), collector,
	)
	if !fieldValid || !directionValid || len(collector.issues) != start {
		return api.RequestSort{}, false
	}
	result, err := api.NewRequestSort(field, direction)
	if err != nil {
		collector.add(path, restqueryv1.IssueValueNotAllowed)
		return api.RequestSort{}, false
	}
	return result, true
}

func decodeDirection(
	raw json.RawMessage, present bool, path string, collector *issueCollector,
) (api.Direction, bool) {
	if !present {
		return 0, false
	}
	value, valid := decodeString(raw, path, collector)
	if !valid {
		return 0, false
	}
	direction, found := api.ParseDirection(value)
	if !found {
		collector.add(path, restqueryv1.IssueValueNotAllowed)
		return 0, false
	}
	return direction, true
}
