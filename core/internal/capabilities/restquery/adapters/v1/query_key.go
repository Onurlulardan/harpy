package v1query

import (
	"strings"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

var simpleQueryKeys = [...]string{
	"page", "limit", "search", "filter_operator",
}

type normalizedLeaf struct {
	segments []string
	value    string
	path     string
	aliasID  string
}

func normalizeQueryLeaves(
	leaves []wireLeaf, manifest manifest, collector *issueCollector,
) []normalizedLeaf {
	result := make([]normalizedLeaf, 0, len(leaves))
	for _, leaf := range leaves {
		segments, valid := splitV1Key(leaf.key)
		if valid && canonicalSegments(segments) {
			result = append(result, normalizedLeaf{
				segments: segments, value: leaf.value, path: leaf.path,
			})
			continue
		}
		binding, found := manifest.queryKeyAliases[leaf.key]
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

func splitV1Key(key string) ([]string, bool) {
	opening := strings.IndexByte(key, '[')
	if opening < 0 {
		return []string{key}, key != ""
	}
	if opening == 0 {
		return nil, false
	}
	segments := []string{key[:opening]}
	remainder := key[opening:]
	for remainder != "" {
		if remainder[0] != '[' {
			return nil, false
		}
		closing := strings.IndexByte(remainder, ']')
		if closing < 0 {
			return nil, false
		}
		segments = append(segments, remainder[1:closing])
		remainder = remainder[closing+1:]
	}
	return segments, true
}

func canonicalSegments(segments []string) bool {
	if len(segments) == 1 {
		return canonicalSimpleKey(segments[0])
	}
	if len(segments) == 2 && segments[0] == "sort" {
		return segments[1] == "field" || segments[1] == "order"
	}
	if len(segments) == 2 && segments[0] == "filters" {
		return segments[1] != ""
	}
	if len(segments) == 3 && segments[0] == "filters" && segments[1] != "" {
		return segments[2] == "operator" || segments[2] == "value" ||
			segments[2] == ""
	}
	return false
}

func canonicalSimpleKey(value string) bool {
	for _, candidate := range simpleQueryKeys {
		if value == candidate {
			return true
		}
	}
	return false
}

func canonicalQueryKey(value string) bool {
	segments, valid := splitV1Key(value)
	return valid && canonicalSegments(segments)
}
