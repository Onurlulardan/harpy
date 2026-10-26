package iam

import (
	"encoding/json"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

type profileSchemaPublishDocument struct {
	SchemaID               string                            `json:"schema_id"`
	Version                int64                             `json:"version"`
	ExpectedActiveSchemaID string                            `json:"expected_active_schema_id"`
	Fields                 []profileFieldDocument            `json:"fields"`
	MigrationMappings      []profileMigrationMappingDocument `json:"migration_mappings,omitempty"`
}

type profileMigrationMappingDocument struct {
	SourceFieldID string `json:"source_field_id"`
	TargetFieldID string `json:"target_field_id"`
	ConverterKind string `json:"converter_kind"`
}

type profileFieldDocument struct {
	ID             string                    `json:"id"`
	Key            string                    `json:"key"`
	Type           string                    `json:"type"`
	Cardinality    string                    `json:"cardinality"`
	Required       bool                      `json:"required"`
	Classification string                    `json:"classification"`
	Normalization  string                    `json:"normalization"`
	State          string                    `json:"state"`
	Validation     profileValidationDocument `json:"validation"`
}

type profileValidationDocument struct {
	MaximumBytes     int          `json:"maximum_bytes"`
	MaximumItems     int          `json:"maximum_items,omitempty"`
	IntegerMinimum   *json.Number `json:"integer_minimum,omitempty"`
	IntegerMaximum   *json.Number `json:"integer_maximum,omitempty"`
	DecimalPrecision int          `json:"decimal_precision,omitempty"`
	DecimalScale     int          `json:"decimal_scale,omitempty"`
	EnumValues       []string     `json:"enum_values,omitempty"`
}

type profileSchemaRetireDocument struct {
	SchemaID               string `json:"schema_id"`
	ExpectedActiveSchemaID string `json:"expected_active_schema_id"`
	ExpectedFieldRevision  int64  `json:"expected_field_revision"`
}

func profilePublishInput(
	resource profileapp.ResourceKind,
	key string,
	document profileSchemaPublishDocument,
) (ProfileSchemaPublish, bool) {
	schemaID, schemaValid := decodeUUID(document.SchemaID)
	activeID, activeValid := decodeUUID(document.ExpectedActiveSchemaID)
	if !schemaValid || !activeValid || document.Version < 2 || document.Fields == nil {
		return ProfileSchemaPublish{}, false
	}
	fields := make([]ProfileFieldInput, len(document.Fields))
	for index, candidate := range document.Fields {
		field, valid := profileFieldInput(candidate)
		if !valid {
			return ProfileSchemaPublish{}, false
		}
		fields[index] = field
	}
	mappings := make([]ProfileMigrationMappingInput, len(document.MigrationMappings))
	for index, candidate := range document.MigrationMappings {
		sourceID, sourceValid := decodeUUID(candidate.SourceFieldID)
		targetID, targetValid := decodeUUID(candidate.TargetFieldID)
		if !sourceValid || !targetValid {
			return ProfileSchemaPublish{}, false
		}
		mappings[index] = ProfileMigrationMappingInput{
			SourceFieldID: sourceID, TargetFieldID: targetID,
			ConverterKind: candidate.ConverterKind,
		}
	}
	mappings, valid := canonicalProfileMigrationMappings(mappings)
	if !valid {
		return ProfileSchemaPublish{}, false
	}
	return ProfileSchemaPublish{
		IdempotencyKey: key, SchemaID: schemaID, Resource: resource,
		Version: document.Version, ExpectedActiveSchemaID: activeID, Fields: fields,
		MigrationMappings: mappings,
	}, true
}

func profileFieldInput(document profileFieldDocument) (ProfileFieldInput, bool) {
	id, valid := decodeUUID(document.ID)
	if !valid {
		return ProfileFieldInput{}, false
	}
	validation := ProfileValidationInput{
		MaximumBytes:     document.Validation.MaximumBytes,
		MaximumItems:     document.Validation.MaximumItems,
		DecimalPrecision: document.Validation.DecimalPrecision,
		DecimalScale:     document.Validation.DecimalScale,
		EnumValues:       document.Validation.EnumValues,
	}
	if document.Validation.IntegerMinimum != nil {
		value := document.Validation.IntegerMinimum.String()
		validation.IntegerMinimum = &value
	}
	if document.Validation.IntegerMaximum != nil {
		value := document.Validation.IntegerMaximum.String()
		validation.IntegerMaximum = &value
	}
	return ProfileFieldInput{
		ID: id, Key: document.Key, Type: profileapp.LogicalType(document.Type),
		Cardinality:    profileapp.Cardinality(document.Cardinality),
		Required:       document.Required,
		Classification: profileapp.Classification(document.Classification),
		Normalization:  profileapp.NormalizationProfile(document.Normalization),
		State:          profileapp.FieldState(document.State), Validation: validation,
	}, true
}
