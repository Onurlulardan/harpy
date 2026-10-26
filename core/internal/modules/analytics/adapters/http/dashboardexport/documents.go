package dashboardexporthttp

import "encoding/json"

type submitDocument struct {
	Format     string          `json:"format"`
	Parameters json.RawMessage `json:"parameters,omitempty"`
}
