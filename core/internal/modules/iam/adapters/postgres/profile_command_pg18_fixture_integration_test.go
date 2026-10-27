//go:build integration

package postgres

import (
	"fmt"
	"testing"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

const (
	iamProfileCommandSchemaV2ID = "d1000000-0000-4000-8000-000000000001"
	iamProfileCommandSchemaV3ID = "d1000000-0000-4000-8000-000000000002"
	iamProfileCommandFieldID    = "d2000000-0000-4000-8000-000000000001"
)

type iamProfileCommandFixture struct {
	published profileapp.Schema
	retired   profileapp.Schema
	document  profileapp.Document
}

func newIAMProfileCommandFixture(
	t *testing.T,
	maximumBytes int,
	value string,
) iamProfileCommandFixture {
	t.Helper()
	active := newIAMProfileCommandField(t, profileapp.FieldActive, maximumBytes)
	retired := newIAMProfileCommandField(t, profileapp.FieldRetired, maximumBytes)
	published := newIAMProfileCommandSchema(
		t, iamProfileCommandSchemaV2ID, 2, active,
	)
	retiredSchema := newIAMProfileCommandSchema(
		t, iamProfileCommandSchemaV3ID, 3, retired,
	)
	documentJSON := []byte(fmt.Sprintf(
		`{"%s":%q}`, iamProfileCommandFieldID, value,
	))
	document, err := profileapp.DecodeDocument(published, documentJSON, nil, nil)
	if err != nil {
		t.Fatalf("create profile command document: %v", err)
	}
	return iamProfileCommandFixture{
		published: published, retired: retiredSchema, document: document,
	}
}

func newIAMProfileCommandField(
	t *testing.T,
	state profileapp.FieldState,
	maximumBytes int,
) profileapp.Field {
	t.Helper()
	field, err := profileapp.NewField(profileapp.FieldInput{
		ID:       uuid.MustParse(iamProfileCommandFieldID),
		Resource: profileapp.ResourceUser, Key: "display_name",
		Type: profileapp.LogicalString, Cardinality: profileapp.CardinalitySingle,
		Classification: profileapp.ClassificationInternal,
		Normalization:  profileapp.NormalizationNoneV1, State: state,
		Validation: profileapp.ValidationInput{MaximumBytes: maximumBytes},
	})
	if err != nil {
		t.Fatalf("create profile command field: %v", err)
	}
	return field
}

func newIAMProfileCommandSchema(
	t *testing.T,
	id string,
	version int64,
	field profileapp.Field,
) profileapp.Schema {
	t.Helper()
	schema, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: uuid.MustParse(id), Resource: profileapp.ResourceUser,
		Version: version, State: profileapp.SchemaActive,
		Fields: []profileapp.Field{field},
	})
	if err != nil {
		t.Fatalf("create profile command schema: %v", err)
	}
	return schema
}
