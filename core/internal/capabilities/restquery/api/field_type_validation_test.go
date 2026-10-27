package api

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestTextListFieldTypeRequiresExactTextProfile(t *testing.T) {
	exact := mustFieldTextProfile(t, TextCasefoldDisabled, TextAccentDisabled)
	companion := mustFieldTextProfile(
		t, TextCasefoldUnicodeDefaultV1, TextAccentInsensitiveV1,
	)
	fieldType, err := NewTextListFieldType(exact)
	if err != nil {
		t.Fatal(err)
	}
	profile, found, accessorErr := fieldType.Text()
	if accessorErr != nil || !found || profile.Validate() != nil {
		t.Fatal("exact text-list profile unavailable")
	}
	if _, err := NewTextListFieldType(companion); err == nil {
		t.Fatal("text-list accepted companion semantics")
	}
	if _, err := NewTextFieldType(companion); err != nil {
		t.Fatal("scalar text rejected allowed companion semantics")
	}
}

func TestFieldTypeRejectsInvalidProfilesAndStateTampering(t *testing.T) {
	if _, err := NewTextFieldType(TextProfile{}); err == nil {
		t.Fatal("invalid text profile accepted")
	}
	if _, err := NewIntegerFieldType(IntegerProfile{}); err == nil {
		t.Fatal("invalid integer profile accepted")
	}
	if _, err := NewDecimalFieldType(DecimalProfile{}); err == nil {
		t.Fatal("invalid decimal profile accepted")
	}
	if _, err := NewEnumFieldType(EnumProfile{}); err == nil {
		t.Fatal("invalid enum profile accepted")
	}
	text := mustFieldTextProfile(t, TextCasefoldDisabled, TextAccentDisabled)
	integer := mustFieldIntegerProfile(t, 0, 1)
	invalid := []FieldType{
		{},
		{kind: LogicalTypeText, state: textFieldState{profile: text}},
		{kind: LogicalTypeText, state: integerFieldState{profile: integer}, validated: true},
		{kind: LogicalTypeInteger, state: fixedFieldState{kind: LogicalTypeInteger}, validated: true},
		{kind: LogicalTypeBoolean, state: fixedFieldState{kind: LogicalTypeDate}, validated: true},
		{kind: LogicalType(255), state: fixedFieldState{kind: LogicalType(255)}, validated: true},
	}
	for _, fieldType := range invalid {
		if err := fieldType.Validate(); err == nil {
			t.Fatal("invalid field type validated")
		}
		assertFieldTypeAccessorsFail(t, fieldType)
	}
}

func TestFieldTypeRedactsAndRejectsDirectJSON(t *testing.T) {
	fieldType, err := NewEnumFieldType(mustFieldEnumProfile(t))
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(fieldType) != queryValueRedaction ||
		fmt.Sprintf("%#v", fieldType) != queryValueRedaction {
		t.Fatal("field type formatting leaked")
	}
	if _, err := json.Marshal(fieldType); err == nil {
		t.Fatal("field type silently marshaled")
	}
}

func assertFieldTypeAccessorsFail(t *testing.T, fieldType FieldType) {
	t.Helper()
	if _, err := fieldType.Kind(); err == nil {
		t.Fatal("invalid kind accessor succeeded")
	}
	if _, _, err := fieldType.Text(); err == nil {
		t.Fatal("invalid text accessor succeeded")
	}
	if _, _, err := fieldType.Integer(); err == nil {
		t.Fatal("invalid integer accessor succeeded")
	}
	if _, _, err := fieldType.Decimal(); err == nil {
		t.Fatal("invalid decimal accessor succeeded")
	}
	if _, _, err := fieldType.Enum(); err == nil {
		t.Fatal("invalid enum accessor succeeded")
	}
}
