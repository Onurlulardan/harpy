package profile

import (
	"bytes"
	"crypto/sha256"

	"github.com/google/uuid"
)

func (field Field) ID() uuid.UUID                  { return field.id }
func (field Field) Resource() ResourceKind         { return field.resource }
func (field Field) Key() string                    { return field.key }
func (field Field) Type() LogicalType              { return field.kind }
func (field Field) Cardinality() Cardinality       { return field.cardinality }
func (field Field) Required() bool                 { return field.required }
func (field Field) Classification() Classification { return field.classification }
func (field Field) Normalization() NormalizationProfile {
	return field.normalization
}
func (field Field) State() FieldState { return field.state }

func (field Field) ValidationJSON() []byte {
	return bytes.Clone(field.validationJSON)
}

func (field Field) ValidationSHA256() [sha256.Size]byte {
	return field.validationSHA
}

func (field Field) SemanticSHA256() [sha256.Size]byte {
	return field.semanticSHA
}

func (field Field) valid() bool {
	return field.validated && field.id != uuid.Nil && field.resource.Valid() &&
		field.kind.Valid() && field.cardinality.Valid() && field.state.Valid() &&
		len(field.validationJSON) > 0
}

func (field Field) clone() Field {
	copy := field
	copy.validation.enumValues = append([]string(nil), field.validation.enumValues...)
	copy.validationJSON = bytes.Clone(field.validationJSON)
	return copy
}
