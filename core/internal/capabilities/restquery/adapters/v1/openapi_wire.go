package v1query

import (
	"strings"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func normalizeOpenAPILeaves(
	leaves []wireLeaf, policy manifest, collector *issueCollector,
) []normalizedLeaf {
	result := make([]normalizedLeaf, 0, len(leaves)*2)
	for _, leaf := range leaves {
		segments, valid := splitV1Key(leaf.key)
		if !valid {
			collector.add(leaf.path, restqueryv1.IssueMemberUnknown)
			continue
		}
		if openAPISimpleOrSort(segments) {
			result = append(result, normalizedLeaf{
				segments: segments, value: leaf.value, path: leaf.path,
			})
			continue
		}
		if len(segments) == 2 && segments[0] == "filter" && segments[1] != "" {
			result = appendOpenAPIFilter(result, segments[1], leaf)
			continue
		}
		binding, found := policy.queryKeyAliases[leaf.key]
		if !found {
			collector.add(leaf.path, restqueryv1.IssueMemberUnknown)
			continue
		}
		result = append(result, normalizedLeaf{
			segments: []string{binding.key}, value: leaf.value,
			path: leaf.path, aliasID: binding.aliasID,
		})
	}
	return result
}

func openAPISimpleOrSort(segments []string) bool {
	if len(segments) == 1 {
		return segments[0] == "page" || segments[0] == "limit" ||
			segments[0] == "search"
	}
	return len(segments) == 2 && segments[0] == "sort" &&
		(segments[1] == "field" || segments[1] == "order")
}

func appendOpenAPIFilter(
	result []normalizedLeaf, field string, leaf wireLeaf,
) []normalizedLeaf {
	if !strings.Contains(leaf.value, ",") {
		return append(result, normalizedLeaf{
			segments: []string{"filters", field}, value: leaf.value, path: leaf.path,
		})
	}
	return append(result,
		normalizedLeaf{segments: []string{"filters", field, "operator"},
			value: "in", path: leaf.path},
		normalizedLeaf{segments: []string{"filters", field, "value"},
			value: leaf.value, path: leaf.path},
	)
}
