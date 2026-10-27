package profile

import (
	"testing"

	"github.com/google/uuid"
)

func TestProjectLineageDocumentPreservesRetiredCompatibilityValues(t *testing.T) {
	field := mustField(t, baseFieldInput(fieldOneID, "legacy_name", LogicalString))
	source, err := NewPublishedSchema(SchemaInput{
		ID: schemaOneID, Resource: ResourceUser, Version: 1,
		State: SchemaRetired, Fields: []Field{field},
	})
	if err != nil {
		t.Fatal(err)
	}
	retiredInput := baseFieldInput(fieldOneID, "legacy_name", LogicalString)
	retiredInput.State = FieldRetired
	retired := mustField(t, retiredInput)
	active := mustField(t, baseFieldInput(
		uuid.MustParse("55555555-5555-4555-8555-555555555555"),
		"display_name", LogicalString,
	))
	target := mustSchema(t, schemaTwoID, 2, retired, active)
	projected, err := ProjectLineageDocument(
		source, target,
		[]byte(`{"11111111-1111-4111-8111-111111111111":"Ada"}`),
		nil, fixtureTextNormalizer{},
	)
	if err != nil || string(projected.CanonicalJSON()) !=
		`{"11111111-1111-4111-8111-111111111111":"Ada"}` {
		t.Fatalf("lineage projection = %s/%v", projected.CanonicalJSON(), err)
	}
}

func TestProjectLineageDocumentRejectsDroppedValue(t *testing.T) {
	field := mustField(t, baseFieldInput(fieldOneID, "legacy_name", LogicalString))
	source := mustSchema(t, schemaOneID, 1, field)
	replacement := mustField(t, baseFieldInput(
		fieldTwoID, "display_name", LogicalString,
	))
	target := mustSchema(t, schemaTwoID, 2, replacement)
	if _, err := ProjectLineageDocument(
		source, target,
		[]byte(`{"11111111-1111-4111-8111-111111111111":"Ada"}`),
		nil, fixtureTextNormalizer{},
	); err == nil {
		t.Fatal("dropped lineage value was accepted")
	}
}
