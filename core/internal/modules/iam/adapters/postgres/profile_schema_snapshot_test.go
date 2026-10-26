package postgres

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func TestProfileSchemaSnapshotHydratesExactHistoricalContract(t *testing.T) {
	fieldID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	schemaID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	field, err := profileapp.NewField(profileapp.FieldInput{
		ID: fieldID, Resource: profileapp.ResourceUser, Key: "nickname",
		Type: profileapp.LogicalString, Cardinality: profileapp.CardinalitySingle,
		Classification: profileapp.ClassificationInternal,
		Normalization:  profileapp.NormalizationNoneV1, State: profileapp.FieldActive,
	})
	if err != nil {
		t.Fatal(err)
	}
	schema, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: schemaID, Resource: profileapp.ResourceUser, Version: 1,
		State: profileapp.SchemaActive, Fields: []profileapp.Field{field},
	})
	if err != nil {
		t.Fatal(err)
	}
	validation, semantic := field.ValidationSHA256(), field.SemanticSHA256()
	definition := schema.DefinitionSHA256()
	wire := profileSchemaSnapshotWire{
		SchemaID: schemaID.String(), Resource: string(profileapp.ResourceUser),
		Version: 1, State: string(profileapp.SchemaActive),
		DefinitionSHA256: hex.EncodeToString(definition[:]),
		Fields: []profileFieldSnapshotWire{{
			ID: fieldID.String(), Key: field.Key(), Type: string(field.Type()),
			Cardinality: string(field.Cardinality()), Required: field.Required(),
			Classification: string(field.Classification()),
			Normalization:  string(field.Normalization()), State: string(field.State()),
			Validation:       field.ValidationJSON(),
			ValidationSHA256: hex.EncodeToString(validation[:]),
			SemanticSHA256:   hex.EncodeToString(semantic[:]),
		}},
	}
	document, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	hydrated, err := decodeProfileSchemaSnapshot(document)
	if err != nil || hydrated.ID() != schemaID ||
		hydrated.DefinitionSHA256() != definition {
		t.Fatalf("hydrate profile schema snapshot: %v", err)
	}

	wire.Fields[0].Key = "changed_key"
	drifted, _ := json.Marshal(wire)
	if _, err := decodeProfileSchemaSnapshot(drifted); err == nil {
		t.Fatal("profile semantic drift was accepted")
	}
}
