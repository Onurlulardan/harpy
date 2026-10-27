package rowdecode

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func TestDecodeTextBooleanAndNullProduceCanonicalCursorValues(t *testing.T) {
	text, err := Decode(textType(t), "Caf\u00e9")
	if err != nil {
		t.Fatal(err)
	}
	requireResponse(t, text, "\"Caf\u00e9\"")
	last, present, err := text.LastValue()
	if err != nil || !present {
		t.Fatalf("last value missing: %v", err)
	}
	resolvedValue, present, err := last.Value()
	if err != nil || !present {
		t.Fatalf("resolved cursor value missing: %v", err)
	}
	scalar, present, err := resolvedValue.Scalar()
	if err != nil || !present || scalar != "Caf\u00e9" {
		t.Fatalf("cursor scalar = %q, %v, %v", scalar, present, err)
	}

	booleanType, _ := api.NewBooleanFieldType()
	boolean, err := Decode(booleanType, true)
	if err != nil {
		t.Fatal(err)
	}
	requireResponse(t, boolean, "true")

	nullValue, err := Decode(textType(t), nil)
	if err != nil {
		t.Fatal(err)
	}
	requireResponse(t, nullValue, "null")
	nullLast, present, err := nullValue.LastValue()
	if err != nil || !present {
		t.Fatalf("null cursor missing: %v", err)
	}
	isNull, err := nullLast.IsNull()
	if err != nil || !isNull {
		t.Fatalf("cursor null = %v, %v", isNull, err)
	}
}

func TestDecodeScalarRejectsShapeAndCanonicalDrift(t *testing.T) {
	booleanType, _ := api.NewBooleanFieldType()
	cases := []struct {
		name      string
		fieldType api.FieldType
		value     any
	}{
		{"text bytes", textType(t), []byte("secret")},
		{"text non NFC", textType(t), "e\u0301"},
		{"text NUL", textType(t), "a\x00b"},
		{"boolean text", booleanType, "true"},
		{"invalid field", api.FieldType{}, true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Decode(test.fieldType, test.value); err == nil {
				t.Fatal("invalid scalar accepted")
			}
		})
	}
}

func TestDecodeEnumRequiresPublishedStableCode(t *testing.T) {
	first, _ := api.NewEnumValue("OPEN", nil)
	second, _ := api.NewEnumValue("DONE", nil)
	profile, err := api.NewEnumProfile(uuid.New().String(), []api.EnumValue{first, second})
	if err != nil {
		t.Fatal(err)
	}
	fieldType, err := api.NewEnumFieldType(profile)
	if err != nil {
		t.Fatal(err)
	}
	value, err := Decode(fieldType, "DONE")
	if err != nil {
		t.Fatal(err)
	}
	requireResponse(t, value, `"DONE"`)
	for _, invalid := range []any{"UNKNOWN", []byte("OPEN"), "e\u0301"} {
		if _, err := Decode(fieldType, invalid); err == nil {
			t.Fatal("undeclared or wrong-shape enum accepted")
		}
	}
}
