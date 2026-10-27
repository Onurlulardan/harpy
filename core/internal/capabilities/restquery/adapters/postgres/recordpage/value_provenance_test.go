package recordpage

import (
	"math/big"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/rowdecode"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestValidateRowRejectsDecimalProfileTransplant(t *testing.T) {
	expected := provenanceDecimalType(t, 12, 2)
	transplanted := provenanceDecimalType(t, 12, 4)
	value, err := rowdecode.Decode(transplanted, pgtype.Numeric{
		Int: big.NewInt(123400), Exp: -4, Valid: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	column := columnMetadata{fieldType: expected, kind: api.LogicalTypeDecimal}
	if validateRow([]rowdecode.Value{value}, []columnMetadata{column}) != ErrContract {
		t.Fatal("decimal value from a different precision/scale profile was accepted")
	}
}

func TestValidateRowRejectsEnumProfileTransplant(t *testing.T) {
	expected := provenanceEnumType(t, "30000000-0000-4000-8000-000000000001")
	transplanted := provenanceEnumType(t, "30000000-0000-4000-8000-000000000002")
	value, err := rowdecode.Decode(transplanted, "OPEN")
	if err != nil {
		t.Fatal(err)
	}
	column := columnMetadata{fieldType: expected, kind: api.LogicalTypeEnum}
	if validateRow([]rowdecode.Value{value}, []columnMetadata{column}) != ErrContract {
		t.Fatal("enum value from a different published profile was accepted")
	}
	valid, _ := rowdecode.Decode(expected, "OPEN")
	if validateRow([]rowdecode.Value{valid}, []columnMetadata{column}) != nil {
		t.Fatal("value from the exact enum profile was rejected")
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
