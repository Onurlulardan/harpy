package postgres

import (
	"encoding/json"
	"time"
)

type revisionWire struct {
	ID               string          `json:"id"`
	ResourceKind     string          `json:"resource_kind"`
	Version          int64           `json:"version"`
	State            string          `json:"state"`
	Definition       json.RawMessage `json:"definition"`
	DefinitionSHA256 string          `json:"definition_sha256"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	PublishedAt      *time.Time      `json:"published_at"`
	RowRevision      int64           `json:"row_revision"`
}
