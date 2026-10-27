package profile

import (
	"testing"

	"github.com/google/uuid"
)

func TestProjectReadTreatsOldSchemaMissingValueAsNull(t *testing.T) {
	existing := mustField(t, baseFieldInput(fieldOneID, "nickname", LogicalString))
	added := mustField(t, baseFieldInput(fieldTwoID, "start_date", LogicalDate))
	oldSchema := mustSchema(t, schemaOneID, 1, existing)
	activeSchema := mustSchema(t, schemaTwoID, 2, existing, added)
	oldDocument := mustDocument(t, oldSchema,
		`{"11111111-1111-4111-8111-111111111111":""}`)

	projection, err := ProjectRead(
		activeSchema, oldDocument, []uuid.UUID{fieldTwoID, fieldOneID},
	)
	if err != nil {
		t.Fatalf("project old document: %v", err)
	}
	want := `{"11111111-1111-4111-8111-111111111111":"",` +
		`"22222222-2222-4222-8222-222222222222":null}`
	if string(projection) != want {
		t.Fatalf("projection = %s", projection)
	}
}

func TestProjectReadRejectsUnknownRetiredAndDuplicateSelection(t *testing.T) {
	active := mustField(t, baseFieldInput(fieldOneID, "nickname", LogicalString))
	retiredInput := baseFieldInput(fieldTwoID, "old_label", LogicalString)
	retiredInput.State = FieldRetired
	retired := mustField(t, retiredInput)
	schema := mustSchema(t, schemaOneID, 1, active, retired)
	document := mustDocument(t, schema, `{}`)
	tests := map[string][]uuid.UUID{
		"unknown":   {fieldThreeID},
		"retired":   {fieldTwoID},
		"duplicate": {fieldOneID, fieldOneID},
	}
	for name, selection := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := ProjectRead(schema, document, selection); err == nil {
				t.Fatal("invalid projection was accepted")
			}
		})
	}
	if _, err := ProjectRead(schema, document, nil); err == nil {
		t.Fatal("nil selection was accepted")
	}
}

func TestProjectCompatibilityReadIgnoresTargetOnlyValuesAndValidationDrift(t *testing.T) {
	sourceInput := baseFieldInput(fieldOneID, "legacy_name", LogicalString)
	sourceInput.Validation.MaximumBytes = 8
	source := mustSchema(t, schemaOneID, 1, mustField(t, sourceInput))
	retainedInput := sourceInput
	retainedInput.State = FieldRetired
	retainedInput.Validation.MaximumBytes = 64
	targetInput := baseFieldInput(fieldTwoID, "display_name", LogicalString)
	targetInput.Validation.MaximumBytes = 64
	target := mustSchema(
		t, schemaTwoID, 2, mustField(t, retainedInput), mustField(t, targetInput),
	)
	document := mustDocument(t, target,
		`{"11111111-1111-4111-8111-111111111111":"long historical value",`+
			`"22222222-2222-4222-8222-222222222222":"long historical value"}`)
	projection, err := ProjectCompatibilityRead(
		source, document, []uuid.UUID{fieldOneID},
	)
	if err != nil || string(projection) !=
		`{"11111111-1111-4111-8111-111111111111":"long historical value"}` {
		t.Fatalf("compatibility projection=%s/%v", projection, err)
	}
	if _, err = ProjectRead(source, document, []uuid.UUID{fieldOneID}); err == nil {
		t.Fatal("strict projection accepted target-only compatibility document")
	}
}
