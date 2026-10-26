package profile

import (
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
)

type FieldInput struct {
	ID             uuid.UUID
	Resource       ResourceKind
	Key            string
	Type           LogicalType
	Cardinality    Cardinality
	Required       bool
	Classification Classification
	Normalization  NormalizationProfile
	State          FieldState
	Validation     ValidationInput
}

type Field struct {
	id             uuid.UUID
	resource       ResourceKind
	key            string
	kind           LogicalType
	cardinality    Cardinality
	required       bool
	classification Classification
	normalization  NormalizationProfile
	state          FieldState
	validation     validation
	validationJSON []byte
	validationSHA  [sha256.Size]byte
	semanticSHA    [sha256.Size]byte
	validated      bool
}

func NewField(input FieldInput) (Field, error) {
	if input.ID == uuid.Nil || input.ID.Variant() != uuid.RFC4122 ||
		!input.Resource.Valid() || !validFieldKey(input.Key) ||
		reservedFieldKey(input.Resource, input.Key) || !input.Type.Valid() ||
		!input.Cardinality.Valid() || !validTypeCardinality(input.Type, input.Cardinality) ||
		!input.Classification.Valid() || !input.Normalization.Valid() ||
		!input.State.Valid() || input.Normalization != NormalizationNoneV1 &&
		input.Type != LogicalString {
		return Field{}, ErrFieldInvalid
	}
	constraints, err := newValidation(input.Type, input.Cardinality, input.Validation)
	if err != nil {
		return Field{}, err
	}
	validationJSON, err := canonicalMarshal(constraints.definition(input.Type), MaximumSchemaBytes)
	if err != nil {
		return Field{}, fmt.Errorf("%w: validation contract", ErrFieldInvalid)
	}
	semanticJSON, err := canonicalMarshal(semanticDefinition(input), MaximumSchemaBytes)
	if err != nil {
		return Field{}, fmt.Errorf("%w: immutable semantic", ErrFieldInvalid)
	}
	field := Field{
		id: input.ID, resource: input.Resource, key: input.Key, kind: input.Type,
		cardinality: input.Cardinality, required: input.Required,
		classification: input.Classification, normalization: input.Normalization,
		state: input.State, validation: constraints,
		validationJSON: validationJSON, validationSHA: canonicalDigest(validationJSON),
		semanticSHA: canonicalDigest(semanticJSON),
		validated:   true,
	}
	return field, nil
}
