package queryexecutionhttp

import (
	"encoding/json"

	queryexecutionapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/queryexecution"
)

type resultView struct {
	DataSourceID string          `json:"data_source_id"`
	Success      bool            `json:"success"`
	Result       json.RawMessage `json:"result,omitempty"`
	ErrorCode    string          `json:"error_code,omitempty"`
}

type batchView struct {
	Results []resultView `json:"results"`
}

func present(result queryexecutionapp.Result) batchView {
	view := batchView{Results: make([]resultView, len(result.Items))}
	for index, item := range result.Items {
		view.Results[index] = resultView{
			DataSourceID: item.DataSourceID,
			Success:      item.ErrorCode == "",
			Result:       append(json.RawMessage(nil), item.Body...),
			ErrorCode:    item.ErrorCode,
		}
	}
	return view
}
