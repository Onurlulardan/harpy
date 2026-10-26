package httpquery

import (
	"slices"
	"strconv"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func decodePublishedGETIndex(
	segment string, path string, collector *publishedIssueCollector,
) (int, bool) {
	if segment == "" || len(segment) > 1 && segment[0] == '0' {
		collector.add(path, restqueryv1.IssueIndexInvalid)
		return 0, false
	}
	for index := range len(segment) {
		if segment[index] < '0' || segment[index] > '9' {
			collector.add(path, restqueryv1.IssueIndexInvalid)
			return 0, false
		}
	}
	parsed, err := strconv.ParseUint(segment, 10, strconv.IntSize)
	if err != nil {
		collector.add(path, restqueryv1.IssueIndexInvalid)
		return 0, false
	}
	return int(parsed), true
}

func orderedPublishedGETIndices(
	values map[int]getMember, path string, collector *publishedIssueCollector,
) ([]int, bool) {
	indices := make([]int, 0, len(values))
	for index := range values {
		indices = append(indices, index)
	}
	slices.Sort(indices)
	for expected, actual := range indices {
		if expected != actual {
			collector.add(
				path+"/"+strconv.Itoa(actual),
				restqueryv1.IssueIndexNonContiguous,
			)
			return indices, false
		}
	}
	return indices, true
}
