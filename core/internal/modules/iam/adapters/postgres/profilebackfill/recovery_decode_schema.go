package profilebackfill

import (
	"encoding/hex"
	"encoding/json"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

func decodeRecoverySchema(document []byte) (profileapp.Schema, error) {
	var wire recoverySchemaWire
	if len(document) == 0 || json.Unmarshal(document, &wire) != nil ||
		wire.Fields == nil {
		return profileapp.Schema{}, profilebackfillapp.ErrUnavailable
	}
	id, err := uuid.Parse(wire.SchemaID)
	if err != nil || id == uuid.Nil {
		return profileapp.Schema{}, profilebackfillapp.ErrUnavailable
	}
	definitionSHA, err := hex.DecodeString(wire.DefinitionSHA256)
	if err != nil || len(definitionSHA) != 32 {
		return profileapp.Schema{}, profilebackfillapp.ErrUnavailable
	}
	fields := make([]profileapp.Field, 0, len(wire.Fields))
	for _, fieldWire := range wire.Fields {
		field, fieldErr := decodeRecoveryField(
			profileapp.ResourceKind(wire.Resource), fieldWire,
		)
		if fieldErr != nil {
			return profileapp.Schema{}, fieldErr
		}
		fields = append(fields, field)
	}
	schema, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: id, Resource: profileapp.ResourceKind(wire.Resource),
		Version: wire.Version, State: profileapp.SchemaState(wire.State),
		Fields: fields, ExpectedDefinitionSHA256: definitionSHA,
	})
	clear(definitionSHA)
	if err != nil {
		return profileapp.Schema{}, profilebackfillapp.ErrUnavailable
	}
	return schema, nil
}

func decodeRecoveryField(
	resource profileapp.ResourceKind,
	wire recoveryFieldWire,
) (profileapp.Field, error) {
	id, idErr := uuid.Parse(wire.ID)
	validationSHA, validationErr := hex.DecodeString(wire.ValidationSHA256)
	semanticSHA, semanticErr := hex.DecodeString(wire.SemanticSHA256)
	if idErr != nil || validationErr != nil || semanticErr != nil ||
		id == uuid.Nil || len(validationSHA) != 32 || len(semanticSHA) != 32 {
		return profileapp.Field{}, profilebackfillapp.ErrUnavailable
	}
	field, err := profileapp.NewFieldFromSnapshot(profileapp.FieldSnapshotInput{
		ID: id, Resource: resource, Key: wire.Key,
		Type:           profileapp.LogicalType(wire.Type),
		Cardinality:    profileapp.Cardinality(wire.Cardinality),
		Required:       wire.Required,
		Classification: profileapp.Classification(wire.Classification),
		Normalization:  profileapp.NormalizationProfile(wire.Normalization),
		State:          profileapp.FieldState(wire.State), ValidationJSON: wire.Validation,
		ExpectedValidationSHA256: validationSHA,
		ExpectedSemanticSHA256:   semanticSHA,
	})
	clear(validationSHA)
	clear(semanticSHA)
	if err != nil {
		return profileapp.Field{}, profilebackfillapp.ErrUnavailable
	}
	return field, nil
}
