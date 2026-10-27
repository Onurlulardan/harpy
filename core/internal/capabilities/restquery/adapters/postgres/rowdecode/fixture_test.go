package rowdecode

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/jackc/pgx/v5/pgtype"
)

func textType(t testing.TB) api.FieldType {
	t.Helper()
	profile, err := api.NewTextProfile(api.TextCasefoldDisabled, api.TextAccentDisabled)
	if err != nil {
		t.Fatal(err)
	}
	fieldType, err := api.NewTextFieldType(profile)
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}

func textListType(t testing.TB) api.FieldType {
	t.Helper()
	profile, err := api.NewTextProfile(api.TextCasefoldDisabled, api.TextAccentDisabled)
	if err != nil {
		t.Fatal(err)
	}
	fieldType, err := api.NewTextListFieldType(profile)
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}

func integerType(t testing.TB, minimum, maximum int64) api.FieldType {
	t.Helper()
	profile, err := api.NewIntegerProfile(minimum, maximum)
	if err != nil {
		t.Fatal(err)
	}
	fieldType, err := api.NewIntegerFieldType(profile)
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}

func decimalType(t testing.TB, precision, scale int) api.FieldType {
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

func requireResponse(t testing.TB, value Value, expected string) {
	t.Helper()
	response, err := value.ResponseJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(response) != expected {
		t.Fatalf("response = %s, want %s", response, expected)
	}
}

func pgxScanValue(t testing.TB, oid uint32, input any, destination any) {
	t.Helper()
	typeMap := pgtype.NewMap()
	encoded, err := typeMap.Encode(oid, pgtype.BinaryFormatCode, input, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := typeMap.Scan(oid, pgtype.BinaryFormatCode, encoded, destination); err != nil {
		t.Fatal(err)
	}
}
