package profile

import (
	"errors"
	"testing"
)

func TestPatchUsesStableIDsAndNullDeletes(t *testing.T) {
	nameInput := baseFieldInput(fieldOneID, "nickname", LogicalString)
	nameInput.Required = true
	name := mustField(t, nameInput)
	refsInput := baseFieldInput(fieldTwoID, "external_refs", LogicalUUID)
	refsInput.Cardinality = CardinalityList
	refsInput.Validation.MaximumItems = 2
	refs := mustField(t, refsInput)
	schema := mustSchema(t, schemaOneID, 1, name, refs)
	current := mustDocument(t, schema,
		`{"11111111-1111-4111-8111-111111111111":"Alice",`+
			`"22222222-2222-4222-8222-222222222222":[]}`)

	patched, err := Patch(schema, current,
		[]byte(`{"11111111-1111-4111-8111-111111111111":"",`+
			`"22222222-2222-4222-8222-222222222222":null}`),
		fixtureTextNormalizer{})
	if err != nil {
		t.Fatalf("patch: %v", err)
	}
	if value, found := patched.Value(fieldOneID); !found || string(value) != `""` {
		t.Fatalf("empty string was not preserved: %s, %v", value, found)
	}
	if _, found := patched.Value(fieldTwoID); found {
		t.Fatal("null patch did not delete the optional key")
	}
	if _, found := current.Value(fieldTwoID); !found {
		t.Fatal("patch mutated the input document")
	}
}

func TestPatchNormalizesTypedValuesBeforeHashing(t *testing.T) {
	emailInput := baseFieldInput(fieldOneID, "contact_address", LogicalString)
	emailInput.Normalization = NormalizationEmailV1
	email := mustField(t, emailInput)
	decimalInput := baseFieldInput(fieldTwoID, "credit_limit", LogicalDecimal)
	decimalInput.Validation = ValidationInput{DecimalPrecision: 6, DecimalScale: 2}
	decimal := mustField(t, decimalInput)
	schema := mustSchema(t, schemaOneID, 1, email, decimal)
	current := mustDocument(t, schema, `{}`)
	patched, err := Patch(schema, current,
		[]byte(`{"11111111-1111-4111-8111-111111111111":" USER@EXAMPLE.COM ",`+
			`"22222222-2222-4222-8222-222222222222":"12.3"}`),
		fixtureTextNormalizer{})
	if err != nil {
		t.Fatalf("patch: %v", err)
	}
	emailValue, _ := patched.Value(fieldOneID)
	decimalValue, _ := patched.Value(fieldTwoID)
	if string(emailValue) != `"user@example.com"` || string(decimalValue) != `"12.30"` {
		t.Fatalf("typed normalization = %s / %s", emailValue, decimalValue)
	}
	digest := patched.SHA256()
	decoded, err := DecodeDocument(
		schema, patched.CanonicalJSON(), digest[:], fixtureTextNormalizer{},
	)
	if err != nil || decoded.SHA256() != digest {
		t.Fatalf("normalized document is not stable: %v", err)
	}
}

func TestPatchRejectsRequiredDeletion(t *testing.T) {
	input := baseFieldInput(fieldOneID, "nickname", LogicalString)
	input.Required = true
	field := mustField(t, input)
	schema := mustSchema(t, schemaOneID, 1, field)
	current := mustDocument(t, schema,
		`{"11111111-1111-4111-8111-111111111111":"Alice"}`)
	_, err := Patch(schema, current,
		[]byte(`{"11111111-1111-4111-8111-111111111111":null}`),
		fixtureTextNormalizer{})
	if !errors.Is(err, ErrRequiredFieldMissing) {
		t.Fatalf("required deletion error = %v", err)
	}
}

func TestPatchDropsValueForRetiredField(t *testing.T) {
	active := mustField(t, baseFieldInput(fieldOneID, "nickname", LogicalString))
	source := mustSchema(t, schemaOneID, 1, active)
	current := mustDocument(
		t, source, `{"11111111-1111-4111-8111-111111111111":"Alice"}`,
	)
	retired, err := RetiredField(active)
	if err != nil {
		t.Fatalf("retire field: %v", err)
	}
	target := mustSchema(t, schemaTwoID, 2, retired)
	patched, err := Patch(target, current, []byte(`{}`), fixtureTextNormalizer{})
	if err != nil {
		t.Fatalf("patch retired field: %v", err)
	}
	if _, found := patched.Value(fieldOneID); found {
		t.Fatal("retired field value remained in the new document")
	}
}
