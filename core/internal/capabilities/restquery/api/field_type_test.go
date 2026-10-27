package api

import "testing"

func TestFieldTypeExplicitConstructorsCoverLogicalTypes(t *testing.T) {
	text := mustFieldTextProfile(t, TextCasefoldDisabled, TextAccentDisabled)
	integer := mustFieldIntegerProfile(t, -10, 10)
	decimal := mustFieldDecimalProfile(t, 18, 4)
	enum := mustFieldEnumProfile(t)
	tests := []struct {
		name string
		want LogicalType
		new  func() (FieldType, error)
	}{
		{"text", LogicalTypeText, func() (FieldType, error) { return NewTextFieldType(text) }},
		{"integer", LogicalTypeInteger, func() (FieldType, error) { return NewIntegerFieldType(integer) }},
		{"decimal", LogicalTypeDecimal, func() (FieldType, error) { return NewDecimalFieldType(decimal) }},
		{"boolean", LogicalTypeBoolean, NewBooleanFieldType},
		{"date", LogicalTypeDate, NewDateFieldType},
		{"timestamp", LogicalTypeTimestamp, NewTimestampFieldType},
		{"uuid", LogicalTypeUUID, NewUUIDFieldType},
		{"enum", LogicalTypeEnum, func() (FieldType, error) { return NewEnumFieldType(enum) }},
		{"relation", LogicalTypeRelation, NewRelationFieldType},
		{"text_list", LogicalTypeTextList, func() (FieldType, error) { return NewTextListFieldType(text) }},
		{"uuid_list", LogicalTypeUUIDList, NewUUIDListFieldType},
		{"json", LogicalTypeJSON, NewJSONFieldType},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fieldType, err := test.new()
			if err != nil || fieldType.Validate() != nil {
				t.Fatalf("constructor failed: %v", err)
			}
			kind, kindErr := fieldType.Kind()
			if kindErr != nil || kind != test.want {
				t.Fatalf("kind=%v err=%v", kind, kindErr)
			}
		})
	}
}

func TestFieldTypeAccessorsAreTypedAndDefensive(t *testing.T) {
	text := mustFieldTextProfile(t, TextCasefoldDisabled, TextAccentDisabled)
	integer := mustFieldIntegerProfile(t, -5, 5)
	decimal := mustFieldDecimalProfile(t, 20, 5)
	enum := mustFieldEnumProfile(t)
	textType, _ := NewTextFieldType(text)
	integerType, _ := NewIntegerFieldType(integer)
	decimalType, _ := NewDecimalFieldType(decimal)
	enumType, _ := NewEnumFieldType(enum)

	if _, found, err := textType.Text(); err != nil || !found {
		t.Fatal("text profile unavailable")
	}
	if _, found, err := integerType.Integer(); err != nil || !found {
		t.Fatal("integer profile unavailable")
	}
	if _, found, err := decimalType.Decimal(); err != nil || !found {
		t.Fatal("decimal profile unavailable")
	}
	if _, found, err := textType.Enum(); err != nil || found {
		t.Fatal("wrong typed accessor did not fail closed")
	}

	enum.values[0].code = "MUTATED"
	returned, found, err := enumType.Enum()
	if err != nil || !found {
		t.Fatal("enum profile unavailable")
	}
	returned.values[0].code = "MUTATED_AGAIN"
	again, found, err := enumType.Enum()
	if err != nil || !found {
		t.Fatal("enum profile unavailable after copy mutation")
	}
	value, found, err := again.Lookup("OPEN")
	if err != nil || !found || enumCode(value) != "OPEN" {
		t.Fatal("enum field type exposed mutable storage")
	}
}

func mustFieldTextProfile(
	t *testing.T, casefold TextCasefoldProfile, accent TextAccentProfile,
) TextProfile {
	t.Helper()
	profile, err := NewTextProfile(casefold, accent)
	if err != nil {
		t.Fatal(err)
	}
	return profile
}

func mustFieldIntegerProfile(t *testing.T, minimum, maximum int64) IntegerProfile {
	t.Helper()
	profile, err := NewIntegerProfile(minimum, maximum)
	if err != nil {
		t.Fatal(err)
	}
	return profile
}

func mustFieldDecimalProfile(t *testing.T, precision, scale int) DecimalProfile {
	t.Helper()
	profile, err := NewDecimalProfile(precision, scale)
	if err != nil {
		t.Fatal(err)
	}
	return profile
}

func mustFieldEnumProfile(t *testing.T) EnumProfile {
	t.Helper()
	profile, err := NewEnumProfile(
		testEnumVersion, []EnumValue{mustEnumValue(t, "OPEN", nil)},
	)
	if err != nil {
		t.Fatal(err)
	}
	return profile
}
