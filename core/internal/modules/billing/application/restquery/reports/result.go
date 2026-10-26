package reports

import (
	"encoding/json"

	queryruntime "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/runtime"
)

type RecordResult struct {
	items      []json.RawMessage
	hasMore    bool
	nextCursor string
	exactCount *int64
}

func newRecordResult(value queryruntime.Result) RecordResult {
	items := value.Items()
	result := RecordResult{
		items: make([]json.RawMessage, len(items)), hasMore: value.HasMore(),
		nextCursor: value.NextCursor(), exactCount: value.ExactCount(),
	}
	for index, item := range items {
		result.items[index] = append(json.RawMessage(nil), item...)
	}
	return result
}

func (result RecordResult) Items() []json.RawMessage {
	items := make([]json.RawMessage, len(result.items))
	for index, item := range result.items {
		items[index] = append(json.RawMessage(nil), item...)
	}
	return items
}

func (result RecordResult) HasMore() bool      { return result.hasMore }
func (result RecordResult) NextCursor() string { return result.nextCursor }

func (result RecordResult) ExactCount() *int64 {
	if result.exactCount == nil {
		return nil
	}
	value := *result.exactCount
	return &value
}
