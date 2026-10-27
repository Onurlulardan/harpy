package httpquery

import (
	"strconv"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func (document *getDocument) addGETSort(
	leaf getLeaf, collector *issueCollector,
) {
	if len(leaf.segments) < 2 {
		collector.add("/sort", restqueryv1.IssueIndexInvalid)
		return
	}
	indexPath := appendJSONPointer("/sort", leaf.segments[1])
	index, valid := decodeGETIndex(leaf.segments[1], indexPath, collector)
	if !valid {
		return
	}
	if len(leaf.segments) < 3 {
		collector.add(indexPath, restqueryv1.IssueValueNotCanonical)
		return
	}
	if len(leaf.segments) != 3 {
		collector.add(leaf.pointer, restqueryv1.IssueValueNotCanonical)
		return
	}
	item := document.sorts[index]
	if item == nil {
		item = &getSortDocument{}
		document.sorts[index] = item
	}
	member := memberFromGETLeaf(leaf)
	switch leaf.segments[2] {
	case "field":
		item.field = member
	case "direction":
		item.direction = member
	default:
		collector.add(leaf.pointer, restqueryv1.IssueMemberUnknown)
	}
}

func decodeGETSorts(
	document getDocument, collector *issueCollector,
) ([]api.RequestSort, bool) {
	indices, valid := orderedGETIndices(document.sorts, "/sort", collector)
	if len(indices) > 5 {
		collector.addBudget("/sort", restqueryv1.IssueSortCountExceeded)
		valid = false
	}
	result := make([]api.RequestSort, 0, len(indices))
	seen := make(map[string]struct{}, len(indices))
	for _, index := range indices {
		path := "/sort/" + strconv.Itoa(index)
		sort, sortValid := decodeGETSort(document.sorts[index], path, collector)
		if !sortValid {
			valid = false
			continue
		}
		if _, duplicate := seen[sort.Field()]; duplicate {
			collector.add(path+"/field", restqueryv1.IssueFieldDuplicate)
			valid = false
			continue
		}
		seen[sort.Field()] = struct{}{}
		result = append(result, sort)
	}
	return result, valid
}

func decodeGETSort(
	item *getSortDocument, path string, collector *issueCollector,
) (api.RequestSort, bool) {
	if item.field == nil {
		collector.add(path+"/field", restqueryv1.IssueMemberRequired)
	}
	if item.direction == nil {
		collector.add(path+"/direction", restqueryv1.IssueMemberRequired)
	}
	if item.field == nil || item.direction == nil {
		return api.RequestSort{}, false
	}
	field, fieldValid := decodeGETFieldName(*item.field, collector)
	direction, found := api.ParseDirection(item.direction.value)
	if !found {
		collector.add(item.direction.pointer, restqueryv1.IssueValueNotAllowed)
	}
	if !fieldValid || !found {
		return api.RequestSort{}, false
	}
	result, err := api.NewRequestSort(field, direction)
	if err != nil {
		collector.add(path, restqueryv1.IssueValueNotAllowed)
		return api.RequestSort{}, false
	}
	return result, true
}
