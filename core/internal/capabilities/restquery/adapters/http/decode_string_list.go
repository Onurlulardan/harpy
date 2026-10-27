package httpquery

import (
	"strconv"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func decodeFieldList(
	raw []byte,
	path string,
	maximum int,
	limitCode restqueryv1.IssueCode,
	collector *issueCollector,
) ([]string, bool) {
	items, valid := decodeArray(raw, path, collector)
	if !valid {
		return nil, false
	}
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	if len(items) > maximum {
		collector.addBudget(path, limitCode)
		valid = false
	}
	for index, item := range items {
		itemPath := rawMemberPath(path, strconv.Itoa(index))
		name, nameValid := decodeFieldName(item, itemPath, collector)
		if !nameValid {
			valid = false
			continue
		}
		if _, duplicate := seen[name]; duplicate {
			collector.add(itemPath, restqueryv1.IssueFieldDuplicate)
			valid = false
			continue
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}
	return result, valid
}
