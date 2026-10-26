package rowdecode

import (
	"math/big"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestLastValueCarriesExactCanonicalScalarForEverySortableType(t *testing.T) {
	dateType, _ := api.NewDateFieldType()
	timestampType, _ := api.NewTimestampFieldType()
	uuidType, _ := api.NewUUIDFieldType()
	relationType, _ := api.NewRelationFieldType()
	booleanType, _ := api.NewBooleanFieldType()
	identifier := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	enumType := cursorEnumType(t)
	tests := []struct {
		name      string
		fieldType api.FieldType
		driver    any
		kind      api.LogicalType
		scalar    string
		boolean   *bool
	}{
		{"text", textType(t), "Caf\u00e9", api.LogicalTypeText, "Caf\u00e9", nil},
		{"integer", integerType(t, -20, 20), int64(-2), api.LogicalTypeInteger, "-2", nil},
		{"decimal", decimalType(t, 5, 2), pgtype.Numeric{Int: big.NewInt(1234), Exp: -2, Valid: true}, api.LogicalTypeDecimal, "12.34", nil},
		{"date", dateType, time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC), api.LogicalTypeDate, "2026-07-15", nil},
		{"timestamp", timestampType, time.Date(2026, 7, 15, 1, 2, 3, 4_000, time.UTC), api.LogicalTypeTimestamp, "2026-07-15T01:02:03.000004Z", nil},
		{"uuid", uuidType, [16]byte(identifier), api.LogicalTypeUUID, identifier.String(), nil},
		{"enum", enumType, "OPEN", api.LogicalTypeEnum, "OPEN", nil},
		{"relation", relationType, [16]byte(identifier), api.LogicalTypeRelation, identifier.String(), nil},
		{"boolean", booleanType, true, api.LogicalTypeBoolean, "", boolPointer(true)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			decoded, err := Decode(test.fieldType, test.driver)
			if err != nil {
				t.Fatal(err)
			}
			last, present, err := decoded.LastValue()
			if err != nil || !present {
				t.Fatalf("last value missing: %v", err)
			}
			value, present, err := last.Value()
			if err != nil || !present {
				t.Fatalf("cursor scalar missing: %v", err)
			}
			kind, err := value.Kind()
			if err != nil || kind != test.kind {
				t.Fatalf("kind = %v, %v", kind, err)
			}
			if test.boolean != nil {
				actual, present, err := value.Boolean()
				if err != nil || !present || actual != *test.boolean {
					t.Fatalf("boolean = %v, %v, %v", actual, present, err)
				}
				return
			}
			actual, present, err := value.Scalar()
			if err != nil || !present || actual != test.scalar {
				t.Fatalf("scalar = %q, %v, %v", actual, present, err)
			}
		})
	}
}

func cursorEnumType(t testing.TB) api.FieldType {
	t.Helper()
	ordinal := 1
	value, _ := api.NewEnumValue("OPEN", &ordinal)
	profile, _ := api.NewEnumProfile(uuid.New().String(), []api.EnumValue{value})
	fieldType, err := api.NewEnumFieldType(profile)
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}

func boolPointer(value bool) *bool { return &value }
