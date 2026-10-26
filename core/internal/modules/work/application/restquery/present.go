package restquery

import (
	"encoding/json"

	queryruntime "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/runtime"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func Document(result queryruntime.Result) (json.RawMessage, error) {
	items := result.Items()
	values := make([]json.RawMessage, len(items))
	for index, item := range items {
		if !json.Valid(item) {
			return nil, ports.ErrUnavailable
		}
		values[index] = append(json.RawMessage(nil), item...)
	}
	page := struct {
		Size       int    `json:"size"`
		HasMore    bool   `json:"has_more"`
		NextCursor string `json:"next_cursor,omitempty"`
	}{len(values), result.HasMore(), result.NextCursor()}
	return json.Marshal(struct {
		Items []json.RawMessage `json:"items"`
		Page  any               `json:"page"`
	}{values, page})
}
