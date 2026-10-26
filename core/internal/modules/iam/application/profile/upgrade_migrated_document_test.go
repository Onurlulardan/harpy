package profile

import (
	"errors"
	"testing"
)

func TestUpgradeMigratedDocumentBypassesUntouchedHistoricalValidationDrift(t *testing.T) {
	historicalInput := baseFieldInput(fieldOneID, "legacy_name", LogicalString)
	historicalInput.Validation.MaximumBytes = 64
	historical := mustField(t, historicalInput)
	sourceInput := historicalInput
	sourceInput.Validation.MaximumBytes = 8
	source := mustSchema(t, schemaOneID, 2, mustField(t, sourceInput))
	retiredInput := historicalInput
	retiredInput.State = FieldRetired
	targetInput := baseFieldInput(fieldTwoID, "display_name", LogicalString)
	targetInput.Validation.MaximumBytes = 64
	target := mustSchema(
		t, schemaTwoID, 3, mustField(t, retiredInput), mustField(t, targetInput),
	)
	mappings := []MigrationMapping{{
		SourceFieldID: fieldOneID, TargetFieldID: fieldTwoID,
		Converter: MigrationNormalize,
	}}
	historicalSchema := mustHistoricalSchema(t, historical)
	legacy := mustDocument(t, historicalSchema,
		`{"11111111-1111-4111-8111-111111111111":"Legacy value beyond eight"}`)
	digest := legacy.SHA256()
	planned, err := MigrateDocument(
		historicalSchema, target, mappings, legacy.CanonicalJSON(), digest[:],
		fixtureTextNormalizer{},
	)
	if err != nil {
		t.Fatalf("historical migration: %v", err)
	}
	patched, err := UpgradeMigratedDocument(
		source, target, mappings, planned, []byte(`{}`), fixtureTextNormalizer{},
	)
	if err != nil || string(patched.CanonicalJSON()) != string(planned.CanonicalJSON()) {
		t.Fatalf("migrated patch=%s/%v", patched.CanonicalJSON(), err)
	}
	if _, err = UpgradeMigratedDocument(source, target, mappings, planned,
		[]byte(`{"11111111-1111-4111-8111-111111111111":"still too long"}`),
		fixtureTextNormalizer{}); err != ErrPatchInvalid {
		t.Fatalf("tightened source validation err=%v", err)
	}
}

func TestUpgradeMigratedDocumentRejectsRequiredSourceShadowDeletion(t *testing.T) {
	source, target, mappings, planned := mustNormalizeMigrationDocument(t, true)
	before := string(planned.CanonicalJSON())

	_, err := UpgradeMigratedDocument(
		source, target, mappings, planned,
		[]byte(`{"11111111-1111-4111-8111-111111111111":null}`),
		fixtureTextNormalizer{},
	)
	if !errors.Is(err, ErrRequiredFieldMissing) {
		t.Fatalf("required source deletion error = %v", err)
	}
	if string(planned.CanonicalJSON()) != before {
		t.Fatal("failed patch mutated the current target document")
	}
	if _, found := planned.Value(fieldOneID); !found {
		t.Fatal("failed patch removed the required source shadow")
	}
	if _, found := planned.Value(fieldTwoID); !found {
		t.Fatal("failed patch removed the normalized target value")
	}
}

func TestUpgradeMigratedDocumentDeletesOptionalSourceAndTargetValues(t *testing.T) {
	source, target, mappings, planned := mustNormalizeMigrationDocument(t, false)

	patched, err := UpgradeMigratedDocument(
		source, target, mappings, planned,
		[]byte(`{"11111111-1111-4111-8111-111111111111":null}`),
		fixtureTextNormalizer{},
	)
	if err != nil {
		t.Fatalf("optional source deletion: %v", err)
	}
	if _, found := patched.Value(fieldOneID); found {
		t.Fatal("optional source shadow remained after deletion")
	}
	if _, found := patched.Value(fieldTwoID); found {
		t.Fatal("optional normalized target remained after deletion")
	}
}

func mustNormalizeMigrationDocument(
	t *testing.T, required bool,
) (Schema, Schema, []MigrationMapping, Document) {
	t.Helper()
	sourceInput := baseFieldInput(fieldOneID, "legacy_email", LogicalString)
	sourceInput.Required = required
	sourceField := mustField(t, sourceInput)
	source := mustSchema(t, schemaOneID, 1, sourceField)
	retiredInput := sourceInput
	retiredInput.State = FieldRetired
	targetInput := baseFieldInput(fieldTwoID, "contact_email", LogicalString)
	targetInput.Normalization = NormalizationEmailV1
	target := mustSchema(
		t, schemaTwoID, 2, mustField(t, retiredInput), mustField(t, targetInput),
	)
	mappings := []MigrationMapping{{
		SourceFieldID: fieldOneID, TargetFieldID: fieldTwoID,
		Converter: MigrationNormalize,
	}}
	sourceDocument := mustDocument(
		t, source, `{"11111111-1111-4111-8111-111111111111":"ADA@EXAMPLE.COM"}`,
	)
	digest := sourceDocument.SHA256()
	planned, err := MigrateDocument(
		source, target, mappings, sourceDocument.CanonicalJSON(), digest[:],
		fixtureTextNormalizer{},
	)
	if err != nil {
		t.Fatalf("normalize migration: %v", err)
	}
	return source, target, mappings, planned
}

func mustHistoricalSchema(t *testing.T, field Field) Schema {
	t.Helper()
	schema, err := NewPublishedSchema(SchemaInput{
		ID: schemaOneID, Resource: ResourceUser, Version: 1,
		State: SchemaRetired, Fields: []Field{field},
	})
	if err != nil {
		t.Fatal(err)
	}
	return schema
}
