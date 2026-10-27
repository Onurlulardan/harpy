package profilepostgres

import "encoding/json"

type profileSchemaWire struct {
	SchemaID         string             `json:"schema_id"`
	Resource         string             `json:"resource_kind"`
	Version          int64              `json:"version"`
	State            string             `json:"state"`
	DefinitionSHA256 string             `json:"definition_sha256"`
	Fields           []profileFieldWire `json:"fields"`
}

type profileFieldWire struct {
	ID               string          `json:"id"`
	Key              string          `json:"key"`
	Type             string          `json:"logical_type"`
	Cardinality      string          `json:"cardinality"`
	Required         bool            `json:"required"`
	Classification   string          `json:"classification"`
	Normalization    string          `json:"normalization_profile"`
	State            string          `json:"state"`
	Validation       json.RawMessage `json:"validation"`
	ValidationSHA256 string          `json:"validation_sha256"`
	SemanticSHA256   string          `json:"semantic_sha256"`
}

type profileDocumentWire struct {
	SchemaID string          `json:"schema_id"`
	Values   json.RawMessage `json:"values"`
	SHA256   string          `json:"sha256"`
	Revision int64           `json:"revision,omitempty"`
}
