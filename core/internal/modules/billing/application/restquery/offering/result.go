package offering

import (
	"encoding/json"

	queryruntime "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/runtime"
)

type Result struct {
	items         []json.RawMessage
	hasMore       bool
	nextCursor    string
	nextTransport string
	exactCount    *int64
}

func newResult(value queryruntime.Result, transport string) Result {
	raw := value.Items()
	items := make([]json.RawMessage, len(raw))
	for index, item := range raw {
		items[index] = json.RawMessage(item)
	}
	return Result{
		items: items, hasMore: value.HasMore(), nextCursor: value.NextCursor(),
		nextTransport: transport, exactCount: value.ExactCount(),
	}
}

func (result Result) Items() []json.RawMessage {
	items := make([]json.RawMessage, len(result.items))
	for index, item := range result.items {
		items[index] = append(json.RawMessage(nil), item...)
	}
	return items
}

func (result Result) HasMore() bool         { return result.hasMore }
func (result Result) NextCursor() string    { return result.nextCursor }
func (result Result) NextTransport() string { return result.nextTransport }

func (result Result) ExactCount() *int64 {
	if result.exactCount == nil {
		return nil
	}
	value := *result.exactCount
	return &value
}
