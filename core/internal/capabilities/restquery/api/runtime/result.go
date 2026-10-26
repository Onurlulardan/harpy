package runtime

import "bytes"

type Result struct {
	items      [][]byte
	nextCursor string
	hasMore    bool
	exactCount *int64
}

func newResult(items [][]byte, hasMore bool, next string, exact *int64) Result {
	return Result{
		items: cloneItems(items), hasMore: hasMore,
		nextCursor: next, exactCount: cloneCount(exact),
	}
}

func (result Result) Items() [][]byte { return cloneItems(result.items) }
func (result Result) HasMore() bool   { return result.hasMore }
func (result Result) NextCursor() string {
	return result.nextCursor
}
func (result Result) ExactCount() *int64 { return cloneCount(result.exactCount) }

func cloneItems(items [][]byte) [][]byte {
	result := make([][]byte, len(items))
	for index, item := range items {
		result[index] = bytes.Clone(item)
	}
	return result
}

func cloneCount(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copyOfValue := *value
	return &copyOfValue
}
