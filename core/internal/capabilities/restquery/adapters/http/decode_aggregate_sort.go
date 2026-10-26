package httpquery

import (
	"encoding/json"
	"strconv"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

var aggregateSortMembers = []string{"group", "measure", "direction"}

func decodeAggregateSorts(
	raw json.RawMessage,
	present bool,
	path string,
	groups map[string]struct{},
	measures map[string]struct{},
	collector *issueCollector,
) ([]api.RequestAggregateSort, bool) {
	if !present {
		return []api.RequestAggregateSort{}, true
	}
	items, valid := decodeArray(raw, path, collector)
	if !valid {
		return nil, false
	}
	if len(items) > 3 {
		collector.addBudget(path, restqueryv1.IssueSortCountExceeded)
		valid = false
	}
	result := make([]api.RequestAggregateSort, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for index, item := range items {
		itemPath := rawMemberPath(path, strconv.Itoa(index))
		sort, itemValid := decodeAggregateSort(item, itemPath, groups, measures, collector)
		if !itemValid {
			valid = false
			continue
		}
		key := sort.Reference().String() + "\x00" + sort.Name()
		if _, duplicate := seen[key]; duplicate {
			collector.add(itemPath, restqueryv1.IssueNameDuplicate)
			valid = false
			continue
		}
		seen[key] = struct{}{}
		result = append(result, sort)
	}
	return result, valid
}

func decodeAggregateSort(
	raw []byte,
	path string,
	groups map[string]struct{},
	measures map[string]struct{},
	collector *issueCollector,
) (api.RequestAggregateSort, bool) {
	start := len(collector.issues)
	object, objectValid := decodeObject(raw, path, aggregateSortMembers, collector)
	if !objectValid {
		return api.RequestAggregateSort{}, false
	}
	reference, name, referenceValid := decodeAggregateReference(
		object, path, groups, measures, collector,
	)
	directionRaw, directionPresent := requireMember(object, path, "direction", collector)
	direction, directionValid := decodeDirection(
		directionRaw, directionPresent, rawMemberPath(path, "direction"), collector,
	)
	if !referenceValid || !directionValid || len(collector.issues) != start {
		return api.RequestAggregateSort{}, false
	}
	result, err := api.NewRequestAggregateSort(reference, name, direction)
	if err != nil {
		collector.add(path, restqueryv1.IssueReferenceInvalid)
		return api.RequestAggregateSort{}, false
	}
	return result, true
}

func decodeAggregateReference(
	object rawObject,
	path string,
	groups map[string]struct{},
	measures map[string]struct{},
	collector *issueCollector,
) (api.AggregateReference, string, bool) {
	present := presentMembers(object, []string{"group", "measure"})
	if len(present) == 0 {
		collector.add(path, restqueryv1.IssueMemberRequired)
		return 0, "", false
	}
	selected := present[0]
	for _, name := range present[1:] {
		collector.add(rawMemberPath(path, name), restqueryv1.IssueMemberForbidden)
	}
	return decodeDeclaredReference(selected, object[selected], path, groups, measures, collector)
}
