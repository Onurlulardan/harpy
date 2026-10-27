package rowdecode

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestValueBindsExactFieldTypeForSQLNull(t *testing.T) {
	expected := provenanceDecimalType(t, 18, 4)
	value, err := Decode(expected, nil)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := value.FieldType()
	exact, exactErr := actual.ExactEqual(expected)
	isNull, nullErr := value.IsNull()
	if err != nil || exactErr != nil || nullErr != nil || !exact || !isNull {
		t.Fatal("SQL NULL lost its exact declared field type")
	}
}

func TestValueFieldTypeAccessorDefensivelyCopiesEnumProfile(t *testing.T) {
	expected := provenanceEnumType(t, "10000000-0000-4000-8000-000000000001")
	value, err := Decode(expected, "OPEN")
	if err != nil {
		t.Fatal(err)
	}
	first, _ := value.FieldType()
	profile, _, _ := first.Enum()
	values, _ := profile.Values()
	replacement, _ := api.NewEnumValue("TAMPERED", nil)
	values[0] = replacement
	second, _ := value.FieldType()
	secondProfile, _, _ := second.Enum()
	secondValues, _ := secondProfile.Values()
	code, _ := secondValues[0].Code()
	if code != "OPEN" {
		t.Fatal("field type accessor exposed enum profile storage")
	}
}

func provenanceDecimalType(t *testing.T, precision, scale int) api.FieldType {
	t.Helper()
	profile, err := api.NewDecimalProfile(precision, scale)
	if err != nil {
		t.Fatal(err)
	}
	fieldType, err := api.NewDecimalFieldType(profile)
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}

func provenanceEnumType(t *testing.T, version string) api.FieldType {
	t.Helper()
	value, _ := api.NewEnumValue("OPEN", nil)
	profile, err := api.NewEnumProfile(version, []api.EnumValue{value})
	if err != nil {
		t.Fatal(err)
	}
	fieldType, err := api.NewEnumFieldType(profile)
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}
