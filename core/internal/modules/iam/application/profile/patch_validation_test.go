package profile

import "testing"

func TestPatchRejectsUnknownRetiredAndAmbiguousKeys(t *testing.T) {
	active := mustField(t, baseFieldInput(fieldOneID, "nickname", LogicalString))
	retiredInput := baseFieldInput(fieldTwoID, "old_label", LogicalString)
	retiredInput.State = FieldRetired
	retired := mustField(t, retiredInput)
	schema := mustSchema(t, schemaOneID, 1, active, retired)
	current := mustDocument(t, schema, `{}`)
	tests := map[string]string{
		"unknown":      `{"33333333-3333-4333-8333-333333333333":"x"}`,
		"public alias": `{"nickname":"x"}`,
		"retired":      `{"22222222-2222-4222-8222-222222222222":"x"}`,
		"duplicate": `{"11111111-1111-4111-8111-111111111111":"a",` +
			`"11111111-1111-4111-8111-111111111111":"b"}`,
	}
	for name, patch := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Patch(
				schema, current, []byte(patch), fixtureTextNormalizer{},
			); err == nil {
				t.Fatal("invalid patch was accepted")
			}
		})
	}
}

func TestPatchEnforcesTypeAndSizeConstraints(t *testing.T) {
	shortInput := baseFieldInput(fieldOneID, "short_label", LogicalString)
	shortInput.Validation.MaximumBytes = 5
	short := mustField(t, shortInput)
	refsInput := baseFieldInput(fieldTwoID, "external_refs", LogicalUUID)
	refsInput.Cardinality = CardinalityList
	refsInput.Validation.MaximumItems = 1
	refs := mustField(t, refsInput)
	schema := mustSchema(t, schemaOneID, 1, short, refs)
	current := mustDocument(t, schema, `{}`)
	tests := map[string]string{
		"field bytes": `{"11111111-1111-4111-8111-111111111111":"abcd"}`,
		"list size": `{"22222222-2222-4222-8222-222222222222":[` +
			`"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",` +
			`"bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"]}`,
		"uuid type": `{"22222222-2222-4222-8222-222222222222":["not-a-uuid"]}`,
	}
	for name, patch := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Patch(
				schema, current, []byte(patch), fixtureTextNormalizer{},
			); err == nil {
				t.Fatal("invalid constrained patch was accepted")
			}
		})
	}
}

func TestPatchRejectsStableFieldSemanticDrift(t *testing.T) {
	oldField := mustField(t, baseFieldInput(fieldOneID, "employee_code", LogicalString))
	oldSchema := mustSchema(t, schemaOneID, 1, oldField)
	current := mustDocument(t, oldSchema,
		`{"11111111-1111-4111-8111-111111111111":"5"}`)
	newInput := baseFieldInput(fieldOneID, "employee_code", LogicalInteger)
	newField := mustField(t, newInput)
	newSchema := mustSchema(t, schemaTwoID, 2, newField)
	if _, err := Patch(newSchema, current, []byte(`{}`), fixtureTextNormalizer{}); err == nil {
		t.Fatal("stable field ID was allowed to change immutable semantics")
	}
}
