package work

import (
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func listDocument(items json.RawMessage, limit int, compatibility bool) (json.RawMessage, error) {
	var values []json.RawMessage
	if json.Unmarshal(items, &values) != nil {
		return nil, ports.ErrUnavailable
	}
	if compatibility {
		return json.Marshal(struct {
			Success bool              `json:"success"`
			Data    []json.RawMessage `json:"data"`
		}{true, values})
	}
	return json.Marshal(struct {
		Items []json.RawMessage `json:"items"`
		Page  map[string]any    `json:"page"`
	}{values, map[string]any{"size": len(values), "has_more": len(values) == limit}})
}

func getDocument(items json.RawMessage, resource string,
	compatibility bool,
) (json.RawMessage, error) {
	var values []json.RawMessage
	if json.Unmarshal(items, &values) != nil {
		return nil, ports.ErrUnavailable
	}
	if len(values) != 1 {
		return nil, ports.ErrNotFound
	}
	value := values[0]
	if resource == "my-role" {
		var board struct {
			Role string `json:"my_role"`
		}
		if json.Unmarshal(value, &board) != nil {
			return nil, ports.ErrUnavailable
		}
		value, _ = json.Marshal(map[string]any{"role": board.Role})
	}
	if !compatibility {
		return value, nil
	}
	return json.Marshal(struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}{true, value})
}

func receiptDocument(receipt ports.OperationReceipt,
	compatibility bool,
) (json.RawMessage, error) {
	value, err := json.Marshal(map[string]any{
		"id": receipt.AggregateID.String(), "row_version": receipt.RowVersion})
	if err != nil || !compatibility {
		return value, err
	}
	return json.Marshal(struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}{true, value})
}
