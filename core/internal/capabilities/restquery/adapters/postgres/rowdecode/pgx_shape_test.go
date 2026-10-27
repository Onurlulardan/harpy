package rowdecode

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestDecodeAcceptsExactPGXDefaultBinaryShapes(t *testing.T) {
	identifier := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	dateType, _ := api.NewDateFieldType()
	timestampType, _ := api.NewTimestampFieldType()
	uuidType, _ := api.NewUUIDFieldType()
	textList := textListType(t)
	uuidList, _ := api.NewUUIDListFieldType()
	tests := []struct {
		name      string
		oid       uint32
		input     any
		fieldType api.FieldType
		want      string
	}{
		{"text", pgtype.TextOID, "Caf\u00e9", textType(t), "\"Caf\u00e9\""},
		{"int8", pgtype.Int8OID, int64(12), integerType(t, -20, 20), `"12"`},
		{"numeric", pgtype.NumericOID, pgtype.Numeric{Int: big.NewInt(1234), Exp: -2, Valid: true}, decimalType(t, 5, 2), `"12.34"`},
		{"date", pgtype.DateOID, time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC), dateType, `"2026-07-15"`},
		{"timestamptz", pgtype.TimestamptzOID, time.Date(2026, 7, 15, 1, 2, 3, 4_000, time.UTC), timestampType, `"2026-07-15T01:02:03.000004Z"`},
		{"uuid", pgtype.UUIDOID, [16]byte(identifier), uuidType, `"123e4567-e89b-12d3-a456-426614174000"`},
		{"text array", pgtype.TextArrayOID, []string{"a", "b"}, textList, `["a","b"]`},
		{"uuid array", pgtype.UUIDArrayOID, [][16]byte{[16]byte(identifier)}, uuidList, `["123e4567-e89b-12d3-a456-426614174000"]`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			driverValue := pgxDefaultRoundTrip(t, test.oid, test.input)
			value, err := Decode(test.fieldType, driverValue)
			if err != nil {
				t.Fatalf("decode pgx %T: %v", driverValue, err)
			}
			requireResponse(t, value, test.want)
		})
	}
}

func TestDecodeJSONRequiresPGXRawByteScan(t *testing.T) {
	fieldType, _ := api.NewJSONFieldType()
	typeMap := pgtype.NewMap()
	encoded, err := typeMap.Encode(
		pgtype.JSONBOID, pgtype.BinaryFormatCode,
		json.RawMessage(`{"b":2,"a":1}`), nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	var defaultValue any
	if err := typeMap.Scan(pgtype.JSONBOID, pgtype.BinaryFormatCode, encoded, &defaultValue); err != nil {
		t.Fatal(err)
	}
	if _, err := Decode(fieldType, defaultValue); err == nil {
		t.Fatal("lossy pgx default JSON tree accepted")
	}
	var raw []byte
	if err := typeMap.Scan(pgtype.JSONBOID, pgtype.BinaryFormatCode, encoded, &raw); err != nil {
		t.Fatal(err)
	}
	value, err := Decode(fieldType, raw)
	if err != nil {
		t.Fatal(err)
	}
	requireResponse(t, value, `{"a":1,"b":2}`)
}

func pgxDefaultRoundTrip(t testing.TB, oid uint32, input any) any {
	t.Helper()
	typeMap := pgtype.NewMap()
	encoded, err := typeMap.Encode(oid, pgtype.BinaryFormatCode, input, nil)
	if err != nil {
		t.Fatal(err)
	}
	var result any
	if err := typeMap.Scan(oid, pgtype.BinaryFormatCode, encoded, &result); err != nil {
		t.Fatal(err)
	}
	return result
}
