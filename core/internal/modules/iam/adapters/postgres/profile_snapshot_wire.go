package postgres

import (
	"encoding/json"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

type profileSchemaSnapshotWire struct {
	SchemaID         string                     `json:"schema_id"`
	Resource         string                     `json:"resource_kind"`
	Version          int64                      `json:"version"`
	State            string                     `json:"state"`
	DefinitionSHA256 string                     `json:"definition_sha256"`
	Fields           []profileFieldSnapshotWire `json:"fields"`
}

type profileFieldSnapshotWire struct {
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

func (wire profileFieldSnapshotWire) field(
	resource profileapp.ResourceKind,
) (profileapp.Field, error) {
	id, err := decodeProfileUUID(wire.ID)
	if err != nil {
		return profileapp.Field{}, errProfileSnapshotUnavailable
	}
	validation, err := decodeProfileSHA256(wire.ValidationSHA256)
	if err != nil {
		return profileapp.Field{}, errProfileSnapshotUnavailable
	}
	semantic, err := decodeProfileSHA256(wire.SemanticSHA256)
	if err != nil {
		return profileapp.Field{}, errProfileSnapshotUnavailable
	}
	field, err := profileapp.NewFieldFromSnapshot(profileapp.FieldSnapshotInput{
		ID: id, Resource: resource, Key: wire.Key,
		Type:           profileapp.LogicalType(wire.Type),
		Cardinality:    profileapp.Cardinality(wire.Cardinality),
		Required:       wire.Required,
		Classification: profileapp.Classification(wire.Classification),
		Normalization:  profileapp.NormalizationProfile(wire.Normalization),
		State:          profileapp.FieldState(wire.State), ValidationJSON: wire.Validation,
		ExpectedValidationSHA256: validation, ExpectedSemanticSHA256: semantic,
	})
	if err != nil {
		return profileapp.Field{}, errProfileSnapshotUnavailable
	}
	return field, nil
}
