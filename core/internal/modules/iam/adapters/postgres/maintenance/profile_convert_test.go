package maintenance

import (
	"crypto/sha256"
	"testing"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func TestProfileNormalizeMappingRequiresDifferentCompatibleFields(t *testing.T) {
	sourceID := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	targetID := uuid.MustParse("12000000-0000-4000-8000-000000000002")
	source := maintenanceMappingSchema(t,
		"13000000-0000-4000-8000-000000000001", 1, sourceID, "legacy_name", false)
	sourceField, _ := source.Field(sourceID)
	retained, err := profileapp.RetiredField(sourceField)
	if err != nil {
		t.Fatal(err)
	}
	targetActive := maintenanceMappingField(t, targetID, "display_name", true)
	target, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID:       uuid.MustParse("13000000-0000-4000-8000-000000000002"),
		Resource: profileapp.ResourceUser, Version: 2,
		State:  profileapp.SchemaActive,
		Fields: []profileapp.Field{retained, targetActive},
	})
	if err != nil {
		t.Fatal(err)
	}
	targetField, _ := target.Field(targetID)
	semantic := targetField.SemanticSHA256()
	mapping := profileMapping{
		sourceID: sourceID, targetID: targetID,
		kind: "copy_normalize_v1", semanticSHA: semantic[:],
	}
	compiled, err := compileProfileMappings(target, []profileMapping{mapping})
	if err != nil {
		t.Fatalf("normalize mapping rejected: %v", err)
	}
	document := []byte(`{"12000000-0000-4000-8000-000000000001":"Ada"}`)
	converted, err := convertProfileDocument(source, target, compiled,
		profileResource{document: document, sha: sha256Bytes(document)})
	if err != nil || converted.SchemaID() != target.ID() {
		t.Fatalf("normalize conversion = %s/%v", converted.SchemaID(), err)
	}
	mapping.kind = "copy_v1"
	compiled, err = compileProfileMappings(target, []profileMapping{mapping})
	if err != nil {
		t.Fatalf("compile copy evidence: %v", err)
	}
	if _, err = convertProfileDocument(source, target, compiled,
		profileResource{document: document, sha: sha256Bytes(document)}); err == nil {
		t.Fatal("copy converter accepted different field identities")
	}
}

func maintenanceMappingSchema(
	t *testing.T, schema string, version int64, fieldID uuid.UUID, key string, required bool,
) profileapp.Schema {
	t.Helper()
	field := maintenanceMappingField(t, fieldID, key, required)
	result, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: uuid.MustParse(schema), Resource: profileapp.ResourceUser,
		Version: version, State: profileapp.SchemaActive, Fields: []profileapp.Field{field},
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func maintenanceMappingField(
	t *testing.T, fieldID uuid.UUID, key string, required bool,
) profileapp.Field {
	t.Helper()
	field, err := profileapp.NewField(profileapp.FieldInput{
		ID: fieldID, Resource: profileapp.ResourceUser, Key: key,
		Type: profileapp.LogicalString, Cardinality: profileapp.CardinalitySingle,
		Required: required, Classification: profileapp.ClassificationInternal,
		Normalization: profileapp.NormalizationNoneV1, State: profileapp.FieldActive,
		Validation: profileapp.ValidationInput{MaximumBytes: 128},
	})
	if err != nil {
		t.Fatal(err)
	}
	return field
}

func sha256Bytes(value []byte) []byte {
	digest := sha256.Sum256(value)
	return digest[:]
}
