package httpquery

import (
	"encoding/json"
	"strconv"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

var includeMembers = []string{"relation", "select", "where", "sort", "limit", "include"}

type includeDecodeState struct {
	edges int
}

func decodeIncludes(
	raw json.RawMessage, present bool, path string, collector *issueCollector,
) ([]api.RequestInclude, bool) {
	if !present {
		return []api.RequestInclude{}, true
	}
	state := includeDecodeState{}
	return decodeIncludeArray(raw, path, 1, &state, collector)
}

func decodeIncludeArray(
	raw []byte,
	path string,
	depth int,
	state *includeDecodeState,
	collector *issueCollector,
) ([]api.RequestInclude, bool) {
	items, valid := decodeArray(raw, path, collector)
	if !valid {
		return nil, false
	}
	result := make([]api.RequestInclude, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for index, item := range items {
		itemPath := rawMemberPath(path, strconv.Itoa(index))
		include, itemValid := decodeInclude(item, itemPath, depth, state, collector)
		if !itemValid {
			valid = false
			continue
		}
		if _, duplicate := seen[include.Relation()]; duplicate {
			collector.add(rawMemberPath(itemPath, "relation"), restqueryv1.IssueNameDuplicate)
			valid = false
			continue
		}
		seen[include.Relation()] = struct{}{}
		result = append(result, include)
	}
	return result, valid
}

func decodeInclude(
	raw []byte,
	path string,
	depth int,
	state *includeDecodeState,
	collector *issueCollector,
) (api.RequestInclude, bool) {
	start := len(collector.issues)
	state.edges++
	if depth > 2 {
		collector.addBudget(path, restqueryv1.IssueIncludeDepthExceeded)
	}
	if state.edges > 3 {
		collector.addBudget(path, restqueryv1.IssueIncludeEdgeCountExceeded)
	}
	object, objectValid := decodeObject(raw, path, includeMembers, collector)
	if !objectValid {
		return api.RequestInclude{}, false
	}
	relationRaw, relationPresent := requireMember(object, path, "relation", collector)
	relation, relationValid := "", false
	if relationPresent {
		relation, relationValid = decodeName(
			relationRaw, rawMemberPath(path, "relation"), collector,
		)
	}
	selects, selectValid := decodeIncludeSelect(object, path, collector)
	where, _, whereValid := decodeOptionalFilter(
		object["where"], object["where"] != nil, rawMemberPath(path, "where"), collector,
	)
	sorts, sortValid := decodeSorts(
		object["sort"], object["sort"] != nil, rawMemberPath(path, "sort"), collector,
	)
	limit, limitValid := decodeIncludeLimit(object, path, collector)
	children, childrenValid := decodeIncludeChildren(object, path, depth, state, collector)
	valid := relationValid && selectValid && whereValid && sortValid &&
		limitValid && childrenValid && len(collector.issues) == start
	if !valid {
		return api.RequestInclude{}, false
	}
	result, err := api.NewRequestInclude(api.RequestIncludeInput{
		Relation: relation, Select: selects, Where: where, Sort: sorts,
		Limit: limit, Include: children,
	})
	if err != nil {
		collector.add(path, restqueryv1.IssueValueNotAllowed)
		return api.RequestInclude{}, false
	}
	return result, true
}
