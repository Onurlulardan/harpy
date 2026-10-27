package rowdecode

import (
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestDecodeDateAndTimestampUseCanonicalUTCForms(t *testing.T) {
	dateType, _ := api.NewDateFieldType()
	date := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	value, err := Decode(dateType, date)
	if err != nil {
		t.Fatal(err)
	}
	requireResponse(t, value, `"2026-07-15"`)

	timestampType, _ := api.NewTimestampFieldType()
	zone := time.FixedZone("test", 3*60*60)
	timestamp := time.Date(2026, 7, 15, 12, 3, 4, 123456000, zone)
	value, err = Decode(timestampType, timestamp)
	if err != nil {
		t.Fatal(err)
	}
	requireResponse(t, value, `"2026-07-15T09:03:04.123456Z"`)
}

func TestDecodeTemporalRejectsMalformedDriverValues(t *testing.T) {
	dateType, _ := api.NewDateFieldType()
	timestampType, _ := api.NewTimestampFieldType()
	cases := []struct {
		fieldType api.FieldType
		value     any
	}{
		{dateType, "2026-07-15"},
		{dateType, time.Date(2026, 7, 15, 1, 0, 0, 0, time.UTC)},
		{dateType, time.Date(2026, 7, 15, 0, 0, 0, 0, time.FixedZone("x", 0))},
		{timestampType, "2026-07-15T00:00:00.000000Z"},
		{timestampType, time.Date(2026, 7, 15, 0, 0, 0, 1, time.UTC)},
		{timestampType, time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)},
		{timestampType, pgtype.Infinity},
	}
	for _, test := range cases {
		if _, err := Decode(test.fieldType, test.value); err == nil {
			t.Fatal("invalid temporal value accepted")
		}
	}
}

func TestDecodeUUIDAndRelationRequirePGXBinaryShape(t *testing.T) {
	identifier := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	uuidType, _ := api.NewUUIDFieldType()
	relationType, _ := api.NewRelationFieldType()
	for _, fieldType := range []api.FieldType{uuidType, relationType} {
		value, err := Decode(fieldType, [16]byte(identifier))
		if err != nil {
			t.Fatal(err)
		}
		requireResponse(t, value, `"123e4567-e89b-12d3-a456-426614174000"`)
	}
	invalidVariant := [16]byte(identifier)
	invalidVariant[8] = 0xe0
	for _, invalid := range []any{identifier.String(), identifier, [16]byte{}, invalidVariant} {
		if _, err := Decode(uuidType, invalid); err == nil {
			t.Fatal("invalid UUID shape or value accepted")
		}
	}
}
