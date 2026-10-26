package queryexecutionhttp

import "encoding/json"

type executeDocument struct {
	Queries []queryDocument `json:"queries"`
}

type queryDocument struct {
	DataSourceID string          `json:"data_source_id"`
	QueryID      string          `json:"query_id"`
	VersionID    string          `json:"version_id"`
	Params       json.RawMessage `json:"params"`
}
