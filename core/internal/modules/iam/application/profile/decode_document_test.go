package profile

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"testing"
)

func TestDecodeDocumentCanonicalizesAndVerifiesDigest(t *testing.T) {
	name := mustField(t, baseFieldInput(fieldOneID, "nickname", LogicalString))
	ageInput := baseFieldInput(fieldTwoID, "employee_number", LogicalInteger)
	ageInput.Required = true
	age := mustField(t, ageInput)
	schema := mustSchema(t, schemaOneID, 1, name, age)
	wire := `{"22222222-2222-4222-8222-222222222222":"42",` +
		`"11111111-1111-4111-8111-111111111111":"Caf\u00e9"}`
	document, err := DecodeDocument(schema, []byte(wire), nil, fixtureTextNormalizer{})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := `{"11111111-1111-4111-8111-111111111111":"Café",` +
		`"22222222-2222-4222-8222-222222222222":"42"}`
	if string(document.CanonicalJSON()) != want {
		t.Fatalf("canonical document = %s", document.CanonicalJSON())
	}
	digest := document.SHA256()
	verified, err := DecodeDocument(schema, []byte(wire), digest[:], fixtureTextNormalizer{})
	if err != nil || verified.SHA256() != digest {
		t.Fatalf("digest verification failed: %v", err)
	}
	returned := document.CanonicalJSON()
	returned[0] = '!'
	if bytes.Equal(returned, document.CanonicalJSON()) {
		t.Fatal("canonical document bytes escaped ownership")
	}
}

func TestDecodeDocumentRoundTripsInstallationDisplayNameUTF8(t *testing.T) {
	displayName := mustField(t,
		baseFieldInput(fieldOneID, "display_name", LogicalString))
	schema := mustSchema(t, schemaOneID, 1, displayName)
	values := map[string]string{
		"ascii":   "Local Instance Owner",
		"turkish": "İşlem Sahibi",
		"escapes": "Yerel \"Sahip\" \\ konsol",
		"control": "satır\nsekme\tayraç\u001f",
		"emoji":   "Yerel Sahip 😀",
	}
	for name, value := range values {
		t.Run(name, func(t *testing.T) {
			encoded, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			wire := []byte(`{"` + fieldOneID.String() + `":` +
				string(encoded) + `}`)
			document, err := DecodeDocument(
				schema, wire, nil, fixtureTextNormalizer{},
			)
			if err != nil {
				t.Fatalf("decode installation display name: %v", err)
			}
			if !bytes.Equal(document.CanonicalJSON(), wire) {
				t.Fatalf("canonical = %s, want %s",
					document.CanonicalJSON(), wire)
			}
			wantDigest := sha256.Sum256(wire)
			if document.SHA256() != wantDigest {
				t.Fatal("installation display name digest drifted")
			}
			if _, err := DecodeDocument(
				schema, wire, wantDigest[:], fixtureTextNormalizer{},
			); err != nil {
				t.Fatalf("verify installation display name digest: %v", err)
			}
		})
	}
}

func TestDecodeDocumentRejectsInvalidStoredState(t *testing.T) {
	name := mustField(t, baseFieldInput(fieldOneID, "nickname", LogicalString))
	requiredInput := baseFieldInput(fieldTwoID, "employee_number", LogicalInteger)
	requiredInput.Required = true
	required := mustField(t, requiredInput)
	schema := mustSchema(t, schemaOneID, 1, name, required)
	validRequired := `"22222222-2222-4222-8222-222222222222":"42"`
	tests := map[string]string{
		"unknown key":          `{` + validRequired + `,"33333333-3333-4333-8333-333333333333":"x"}`,
		"public alias key":     `{` + validRequired + `,"nickname":"x"}`,
		"stored null":          `{` + validRequired + `,"11111111-1111-4111-8111-111111111111":null}`,
		"type mismatch":        `{"22222222-2222-4222-8222-222222222222":42}`,
		"noncanonical integer": `{"22222222-2222-4222-8222-222222222222":"01"}`,
		"nonnormalized text":   `{` + validRequired + `,"11111111-1111-4111-8111-111111111111":"Cafe\u0301"}`,
		"required absent":      `{}`,
		"duplicate": `{"22222222-2222-4222-8222-222222222222":"42",` +
			`"22222222-2222-4222-8222-222222222222":"43"}`,
	}
	for name, wire := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeDocument(schema, []byte(wire), nil, fixtureTextNormalizer{}); err == nil {
				t.Fatal("invalid stored document was accepted")
			}
		})
	}
	badDigest := make([]byte, 32)
	if _, err := DecodeDocument(
		schema, []byte(`{"22222222-2222-4222-8222-222222222222":"42"}`),
		badDigest, fixtureTextNormalizer{},
	); err == nil {
		t.Fatal("incorrect stored digest was accepted")
	}
}
