package iam

import (
	"errors"
	"testing"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func TestProfilePublishedSchemaBuildsImmutableView(t *testing.T) {
	minimum, maximum := "-10", "10"
	input := ProfileSchemaPublish{
		SchemaID: uuid.MustParse("74000000-0000-4000-8000-000000000002"),
		Resource: profileapp.ResourceUser, Version: 2,
		ExpectedActiveSchemaID: uuid.MustParse(
			"74000000-0000-4000-8000-000000000001",
		),
		Fields: []ProfileFieldInput{{
			ID:  uuid.MustParse("75000000-0000-4000-8000-000000000001"),
			Key: "score", Type: profileapp.LogicalInteger,
			Cardinality:    profileapp.CardinalitySingle,
			Classification: profileapp.ClassificationInternal,
			Normalization:  profileapp.NormalizationNoneV1,
			State:          profileapp.FieldActive,
			Validation: ProfileValidationInput{
				MaximumBytes: 64, IntegerMinimum: &minimum, IntegerMaximum: &maximum,
			},
		}},
	}
	schema, err := profilePublishedSchema(input)
	if err != nil {
		t.Fatal(err)
	}
	view, err := newProfileSchemaView(schema)
	if err != nil || view.ID != input.SchemaID || view.Version != 2 ||
		len(view.Fields) != 1 || len(view.DefinitionSHA256) != 64 ||
		len(view.Fields[0].SemanticSHA256) != 64 ||
		len(view.Fields[0].ValidationSHA256) != 64 {
		t.Fatalf("schema view=%+v error=%v", view, err)
	}
}

func TestProfilePublishedSchemaRejectsNoncanonicalInteger(t *testing.T) {
	minimum := "01"
	_, err := profileOptionalInt64(&minimum)
	if !errors.Is(err, ErrProfileHTTPInvalid) {
		t.Fatalf("noncanonical integer error=%v", err)
	}
}

func TestProfileRetirementChangeCreatesNewTerminalSnapshot(t *testing.T) {
	activeField := profileHTTPTestField(
		t, uuid.MustParse("76000000-0000-4000-8000-000000000001"),
		"display_name", profileapp.ClassificationInternal,
	)
	active, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID:       uuid.MustParse("77000000-0000-4000-8000-000000000001"),
		Resource: profileapp.ResourceUser, Version: 1,
		State: profileapp.SchemaActive, Fields: []profileapp.Field{activeField},
	})
	if err != nil {
		t.Fatal(err)
	}
	change, err := profileRetirementChange(active, ProfileSchemaRetire{
		SchemaID:               uuid.MustParse("77000000-0000-4000-8000-000000000002"),
		Resource:               profileapp.ResourceUser,
		ExpectedActiveSchemaID: active.ID(), FieldID: activeField.ID(),
		ExpectedFieldRevision: 1,
	})
	field, found := change.Schema.Field(activeField.ID())
	if err != nil || change.Schema.Version() != 2 || !found ||
		field.State() != profileapp.FieldRetired || activeField.State() != profileapp.FieldActive {
		t.Fatalf("retirement change=%+v field=%+v found=%t error=%v", change, field, found, err)
	}
}
