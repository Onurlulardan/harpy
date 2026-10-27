package profile

import (
	"bytes"
	"crypto/sha256"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

type FieldSnapshotInput struct {
	ID                       uuid.UUID
	Resource                 ResourceKind
	Key                      string
	Type                     LogicalType
	Cardinality              Cardinality
	Required                 bool
	Classification           Classification
	Normalization            NormalizationProfile
	State                    FieldState
	ValidationJSON           []byte
	ExpectedValidationSHA256 []byte
	ExpectedSemanticSHA256   []byte
}

func NewFieldFromSnapshot(input FieldSnapshotInput) (Field, error) {
	if len(input.ExpectedValidationSHA256) != sha256.Size ||
		len(input.ExpectedSemanticSHA256) != sha256.Size {
		return Field{}, ErrFieldInvalid
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		input.ValidationJSON, MaximumSchemaBytes,
	)
	if err != nil {
		return Field{}, ErrValidationInvalid
	}
	validation, err := decodeFieldValidation(canonical)
	if err != nil {
		return Field{}, err
	}
	field, err := NewField(FieldInput{
		ID: input.ID, Resource: input.Resource, Key: input.Key, Type: input.Type,
		Cardinality: input.Cardinality, Required: input.Required,
		Classification: input.Classification, Normalization: input.Normalization,
		State: input.State, Validation: validation,
	})
	if err != nil || !bytes.Equal(field.validationJSON, canonical) ||
		!bytes.Equal(field.validationSHA[:], input.ExpectedValidationSHA256) ||
		!bytes.Equal(field.semanticSHA[:], input.ExpectedSemanticSHA256) {
		return Field{}, ErrFieldInvalid
	}
	return field, nil
}
