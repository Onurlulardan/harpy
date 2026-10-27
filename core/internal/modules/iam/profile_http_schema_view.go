package iam

import (
	"encoding/hex"
	"encoding/json"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

func newProfileSchemaView(schema profileapp.Schema) (ProfileSchemaView, error) {
	if schema.ID().Version() == 0 || !schema.Resource().Valid() ||
		!schema.State().Valid() {
		return ProfileSchemaView{}, ErrProfileHTTPUnavailable
	}
	fields := schema.Fields()
	viewFields := make([]ProfileFieldView, len(fields))
	for index, field := range fields {
		validation := field.ValidationJSON()
		if !json.Valid(validation) {
			return ProfileSchemaView{}, ErrProfileHTTPUnavailable
		}
		validationSHA, semanticSHA := field.ValidationSHA256(), field.SemanticSHA256()
		viewFields[index] = ProfileFieldView{
			ID: field.ID(), Key: field.Key(), Type: field.Type(),
			Cardinality: field.Cardinality(), Required: field.Required(),
			Classification: field.Classification(),
			Normalization:  field.Normalization(), State: field.State(),
			Validation:       validation,
			ValidationSHA256: hex.EncodeToString(validationSHA[:]),
			SemanticSHA256:   hex.EncodeToString(semanticSHA[:]),
		}
	}
	definition := schema.DefinitionJSON()
	if !json.Valid(definition) {
		return ProfileSchemaView{}, ErrProfileHTTPUnavailable
	}
	digest := schema.DefinitionSHA256()
	return ProfileSchemaView{
		ID: schema.ID(), Resource: schema.Resource(), Version: schema.Version(),
		State: schema.State(), Fields: viewFields, Definition: definition,
		DefinitionSHA256: hex.EncodeToString(digest[:]),
	}, nil
}
