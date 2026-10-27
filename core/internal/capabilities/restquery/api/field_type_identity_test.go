package api

import "testing"

func TestFieldTypeExactEqualIncludesDecimalAndEnumProfiles(t *testing.T) {
	decimalA := identityDecimalType(t, 12, 2)
	decimalB := identityDecimalType(t, 12, 4)
	if equal, err := decimalA.ExactEqual(decimalA); err != nil || !equal {
		t.Fatal("identical decimal types did not match")
	}
	if equal, err := decimalA.ExactEqual(decimalB); err != nil || equal {
		t.Fatal("different decimal profiles matched")
	}
	enumA := identityEnumType(t, "10000000-0000-4000-8000-000000000001")
	enumB := identityEnumType(t, "10000000-0000-4000-8000-000000000002")
	if equal, err := enumA.ExactEqual(enumB); err != nil || equal {
		t.Fatal("different enum versions matched")
	}
}

func TestFieldTypeCloneDefensivelyCopiesEnumValues(t *testing.T) {
	original := identityEnumType(t, "20000000-0000-4000-8000-000000000001")
	cloned, err := original.Clone()
	if err != nil {
		t.Fatal(err)
	}
	state := cloned.state.(enumFieldState)
	state.profile.values[0].code = "TAMPERED"
	cloned.state = state
	profile := original.state.(enumFieldState).profile
	if profile.values[0].code != "OPEN" {
		t.Fatal("field type clone shared enum storage")
	}
}

func identityDecimalType(t *testing.T, precision, scale int) FieldType {
	t.Helper()
	profile, err := NewDecimalProfile(precision, scale)
	if err != nil {
		t.Fatal(err)
	}
	fieldType, err := NewDecimalFieldType(profile)
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}

func identityEnumType(t *testing.T, version string) FieldType {
	t.Helper()
	ordinal := 1
	value, _ := NewEnumValue("OPEN", &ordinal)
	profile, err := NewEnumProfile(version, []EnumValue{value})
	if err != nil {
		t.Fatal(err)
	}
	fieldType, err := NewEnumFieldType(profile)
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}
