package iam

import (
	"testing"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func profilePublicationSchema(
	t *testing.T, schemaID string, version int64, fieldID uuid.UUID, required bool,
	maximumBytes int,
) profileapp.Schema {
	t.Helper()
	return profilePublicationSchemaFields(t, schemaID, version,
		profilePublicationField(t, fieldID, "display_name", required, maximumBytes,
			profileapp.FieldActive),
	)
}

func profilePublicationField(
	t *testing.T,
	id uuid.UUID,
	key string,
	required bool,
	maximumBytes int,
	state profileapp.FieldState,
) profileapp.Field {
	t.Helper()
	field, err := profileapp.NewField(profileapp.FieldInput{
		ID: id, Resource: profileapp.ResourceUser, Key: key,
		Type: profileapp.LogicalString, Cardinality: profileapp.CardinalitySingle,
		Required: required, Classification: profileapp.ClassificationInternal,
		Normalization: profileapp.NormalizationNoneV1, State: state,
		Validation: profileapp.ValidationInput{MaximumBytes: maximumBytes},
	})
	if err != nil {
		t.Fatal(err)
	}
	return field
}

func profilePublicationSchemaFields(
	t *testing.T,
	schemaID string,
	version int64,
	fields ...profileapp.Field,
) profileapp.Schema {
	t.Helper()
	if fields == nil {
		fields = []profileapp.Field{}
	}
	schema, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: uuid.MustParse(schemaID), Resource: profileapp.ResourceUser,
		Version: version, State: profileapp.SchemaActive, Fields: fields,
	})
	if err != nil {
		t.Fatal(err)
	}
	return schema
}
