package profile

import "encoding/json"

type schemaDefinition struct {
	Contract string            `json:"contract"`
	ID       string            `json:"id"`
	Resource ResourceKind      `json:"resource_kind"`
	Version  int64             `json:"version"`
	Fields   []fieldDefinition `json:"fields"`
}

type fieldDefinition struct {
	ID             string               `json:"id"`
	Key            string               `json:"key"`
	Type           LogicalType          `json:"logical_type"`
	Cardinality    Cardinality          `json:"cardinality"`
	Required       bool                 `json:"required"`
	Classification Classification       `json:"classification"`
	Normalization  NormalizationProfile `json:"normalization_profile"`
	State          FieldState           `json:"state"`
	Validation     json.RawMessage      `json:"validation"`
}

func buildSchemaDefinition(input SchemaInput, fields []Field) schemaDefinition {
	definition := schemaDefinition{
		Contract: "harpy/iam-profile-schema/v1", ID: input.ID.String(),
		Resource: input.Resource, Version: input.Version,
		Fields: make([]fieldDefinition, 0, len(fields)),
	}
	for _, field := range fields {
		definition.Fields = append(definition.Fields, fieldDefinition{
			ID: field.id.String(), Key: field.key, Type: field.kind,
			Cardinality: field.cardinality, Required: field.required,
			Classification: field.classification, Normalization: field.normalization,
			State:      field.state,
			Validation: field.ValidationJSON(),
		})
	}
	return definition
}
