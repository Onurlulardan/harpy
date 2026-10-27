package profilebackfill

import "encoding/json"

type recoverySchemaWire struct {
	SchemaID         string              `json:"schema_id"`
	Resource         string              `json:"resource_kind"`
	Version          int64               `json:"version"`
	State            string              `json:"state"`
	DefinitionSHA256 string              `json:"definition_sha256"`
	Fields           []recoveryFieldWire `json:"fields"`
}

type recoveryFieldWire struct {
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

type recoveryMappingWire struct {
	SourceFieldID string `json:"source_field_id"`
	TargetFieldID string `json:"target_field_id"`
	ConverterKind string `json:"converter_kind"`
	SemanticSHA   string `json:"semantic_sha256"`
}
