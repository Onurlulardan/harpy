package presentation

import (
	"bytes"
	"testing"

	fieldregistry "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/fieldregistry"
)

const (
	testTextFieldID = "a1000000-0000-4000-8000-000000000002"
	testBoolFieldID = "a1000000-0000-4000-8000-000000000004"
)

func TestFieldCatalogBindsStableTypeProfileAndMode(t *testing.T) {
	text := mustField(t, testTextFieldID, fieldregistry.TypeText,
		fieldregistry.ValueTextNFCV1, FieldEditable)
	boolean := mustField(t, testBoolFieldID, fieldregistry.TypeBoolean,
		fieldregistry.ValueBooleanV1, FieldReadOnly)
	first, err := NewFieldCatalog([]Field{boolean, text})
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewFieldCatalog([]Field{text, boolean})
	if err != nil {
		t.Fatal(err)
	}
	equal, err := first.ExactEqual(second)
	if err != nil || !equal {
		t.Fatalf("catalog equality = %v, %v", equal, err)
	}
	encoded, _ := first.CanonicalJSON()
	if !bytes.Contains(encoded, []byte(FieldCatalogContract)) {
		t.Fatalf("catalog contract missing: %s", encoded)
	}
	field, found, err := first.Field(testBoolFieldID)
	if err != nil || !found {
		t.Fatalf("field lookup = %v, %v", found, err)
	}
	mode, _ := field.Mode()
	if mode != FieldReadOnly {
		t.Fatalf("field mode = %s", mode)
	}
}

func TestFieldCatalogRejectsSemanticDrift(t *testing.T) {
	for _, input := range []struct {
		id      string
		typeID  fieldregistry.TypeID
		profile fieldregistry.ValueProfile
		mode    FieldMode
	}{
		{"not-a-uuid", fieldregistry.TypeText, fieldregistry.ValueTextNFCV1, FieldEditable},
		{testTextFieldID, fieldregistry.TypeText, fieldregistry.ValueBooleanV1, FieldEditable},
		{testTextFieldID, "unknown", fieldregistry.ValueTextNFCV1, FieldEditable},
		{testTextFieldID, fieldregistry.TypeText, fieldregistry.ValueTextNFCV1, "writable"},
	} {
		if _, err := NewField(input.id, input.typeID, input.profile, input.mode); err == nil {
			t.Fatalf("accepted invalid field: %+v", input)
		}
	}
	field := mustField(t, testTextFieldID, fieldregistry.TypeText,
		fieldregistry.ValueTextNFCV1, FieldEditable)
	if _, err := NewFieldCatalog([]Field{field, field}); err == nil {
		t.Fatal("accepted duplicate stable field ID")
	}
}

func mustField(
	t *testing.T,
	id string,
	typeID fieldregistry.TypeID,
	profile fieldregistry.ValueProfile,
	mode FieldMode,
) Field {
	t.Helper()
	field, err := NewField(id, typeID, profile, mode)
	if err != nil {
		t.Fatal(err)
	}
	return field
}
