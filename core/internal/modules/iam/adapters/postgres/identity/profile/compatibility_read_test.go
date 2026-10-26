package profilepostgres

import (
	"testing"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	"github.com/google/uuid"
)

func TestProfileSourceReadAllowsHistoricalValidationDrift(t *testing.T) {
	fieldID := uuid.MustParse("11000000-0000-4000-8000-000000000001")
	historicalField := profileReadField(t, fieldID, 64)
	currentField := profileReadField(t, fieldID, 8)
	historical := profileReadSchema(
		t, "12000000-0000-4000-8000-000000000001", 1,
		profileapp.SchemaRetired, historicalField,
	)
	current := profileReadSchema(
		t, "12000000-0000-4000-8000-000000000002", 2,
		profileapp.SchemaActive, currentField,
	)
	wire := []byte(`{"11000000-0000-4000-8000-000000000001":"long historical value"}`)
	document, err := profileapp.DecodeDocument(
		historical, wire, nil, profileNormalizer{},
	)
	if err != nil {
		t.Fatal(err)
	}
	read, projection, err := profileSourceRead(
		profileWriteContract{source: current}, profileWriteResource{
			actualSchemaID: historical.ID(), actualSchema: historical, document: document,
		},
	)
	if err != nil || read.SchemaID() != historical.ID() || string(projection) != string(wire) {
		t.Fatalf("historical read=%s/%s/%v", read.SchemaID(), projection, err)
	}
}

func profileReadField(t *testing.T, id uuid.UUID, maximum int) profileapp.Field {
	t.Helper()
	field, err := profileapp.NewField(profileapp.FieldInput{
		ID: id, Resource: profileapp.ResourceUser, Key: "display_name",
		Type: profileapp.LogicalString, Cardinality: profileapp.CardinalitySingle,
		Classification: profileapp.ClassificationInternal,
		Normalization:  profileapp.NormalizationNoneV1, State: profileapp.FieldActive,
		Validation: profileapp.ValidationInput{MaximumBytes: maximum},
	})
	if err != nil {
		t.Fatal(err)
	}
	return field
}

func profileReadSchema(
	t *testing.T, rawID string, version int64,
	state profileapp.SchemaState, field profileapp.Field,
) profileapp.Schema {
	t.Helper()
	schema, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: uuid.MustParse(rawID), Resource: profileapp.ResourceKind(identitydomain.ProfileUser),
		Version: version, State: state, Fields: []profileapp.Field{field},
	})
	if err != nil {
		t.Fatal(err)
	}
	return schema
}
