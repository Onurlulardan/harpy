package profilepostgres

import (
	"encoding/hex"
	"encoding/json"
	"errors"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

var errProfile = errors.New("IAM identity profile is unavailable")

func decodeProfileSchema(document []byte) (profileapp.Schema, error) {
	var wire profileSchemaWire
	if len(document) == 0 || json.Unmarshal(document, &wire) != nil || wire.Fields == nil {
		return profileapp.Schema{}, errProfile
	}
	id, err := uuid.Parse(wire.SchemaID)
	digest, digestErr := hex.DecodeString(wire.DefinitionSHA256)
	if err != nil || digestErr != nil || len(digest) != 32 {
		return profileapp.Schema{}, errProfile
	}
	fields := make([]profileapp.Field, 0, len(wire.Fields))
	for _, fieldWire := range wire.Fields {
		field, fieldErr := decodeProfileField(fieldWire, profileapp.ResourceKind(wire.Resource))
		if fieldErr != nil {
			return profileapp.Schema{}, errProfile
		}
		fields = append(fields, field)
	}
	schema, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: id, Resource: profileapp.ResourceKind(wire.Resource), Version: wire.Version,
		State: profileapp.SchemaState(wire.State), Fields: fields,
		ExpectedDefinitionSHA256: digest,
	})
	if err != nil {
		return profileapp.Schema{}, errProfile
	}
	return schema, nil
}

func decodeProfileField(wire profileFieldWire, resource profileapp.ResourceKind) (profileapp.Field, error) {
	id, err := uuid.Parse(wire.ID)
	validation, validationErr := hex.DecodeString(wire.ValidationSHA256)
	semantic, semanticErr := hex.DecodeString(wire.SemanticSHA256)
	if err != nil || validationErr != nil || semanticErr != nil ||
		len(validation) != 32 || len(semantic) != 32 {
		return profileapp.Field{}, errProfile
	}
	return profileapp.NewFieldFromSnapshot(profileapp.FieldSnapshotInput{
		ID: id, Resource: resource, Key: wire.Key, Type: profileapp.LogicalType(wire.Type),
		Cardinality: profileapp.Cardinality(wire.Cardinality), Required: wire.Required,
		Classification: profileapp.Classification(wire.Classification),
		Normalization:  profileapp.NormalizationProfile(wire.Normalization),
		State:          wireFieldState(wire), ValidationJSON: wire.Validation,
		ExpectedValidationSHA256: validation, ExpectedSemanticSHA256: semantic,
	})
}

func wireFieldState(wire profileFieldWire) profileapp.FieldState {
	return profileapp.FieldState(wire.State)
}
