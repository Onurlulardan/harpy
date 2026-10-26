package cataloghttp

import (
	"encoding/json"
	"time"
)

type MenuConfig struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Config    json.RawMessage `json:"config"`
	IsActive  bool            `json:"is_active"`
	CreatedBy string          `json:"created_by"`
	UpdatedBy string          `json:"updated_by"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type MenuConfigWriteInput struct {
	Name           string
	Config         json.RawMessage
	Active         bool
	IdempotencyKey string
}
