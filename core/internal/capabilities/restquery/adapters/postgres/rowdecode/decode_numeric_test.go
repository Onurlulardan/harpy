package rowdecode

import (
	"math/big"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestDecodeIntegerAcceptsExactSignedAndIntegralNumericShapes(t *testing.T) {
	fieldType := integerType(t, -20, 20)
	cases := []struct {
		value any
		want  string
	}{
		{int16(-2), `"-2"`},
		{int32(3), `"3"`},
		{int64(4), `"4"`},
		{pgtype.Numeric{Int: big.NewInt(2), Exp: 1, Valid: true}, `"20"`},
		{pgtype.Numeric{Int: big.NewInt(200), Exp: -1, Valid: true}, `"20"`},
	}
	for _, test := range cases {
		value, err := Decode(fieldType, test.value)
		if err != nil {
			t.Fatalf("decode %T: %v", test.value, err)
		}
		requireResponse(t, value, test.want)
	}
}

func TestDecodeIntegerRejectsCoercionRangeAndNonFiniteNumeric(t *testing.T) {
	fieldType := integerType(t, -20, 20)
	cases := []any{
		"1", float64(1), int(1), int64(21),
		pgtype.Numeric{Int: big.NewInt(11), Exp: -1, Valid: true},
		pgtype.Numeric{Int: big.NewInt(1), Exp: 20, Valid: true},
		pgtype.Numeric{Int: big.NewInt(1), NaN: true, Valid: true},
		pgtype.Numeric{Int: big.NewInt(1), InfinityModifier: pgtype.Infinity, Valid: true},
		pgtype.Numeric{Int: big.NewInt(1)},
		pgtype.Numeric{Valid: true},
	}
	for _, candidate := range cases {
		if _, err := Decode(fieldType, candidate); err == nil {
			t.Fatalf("invalid integer %T accepted", candidate)
		}
	}
}

func TestDecodeDecimalPadsDeclaredScaleWithoutRounding(t *testing.T) {
	fieldType := decimalType(t, 5, 2)
	cases := []struct {
		value pgtype.Numeric
		want  string
	}{
		{pgtype.Numeric{Int: big.NewInt(1234), Exp: -2, Valid: true}, `"12.34"`},
		{pgtype.Numeric{Int: big.NewInt(12), Valid: true}, `"12.00"`},
		{pgtype.Numeric{Int: big.NewInt(-5), Valid: true}, `"-5.00"`},
		{pgtype.Numeric{Int: big.NewInt(0), Exp: -30, Valid: true}, `"0.00"`},
		{pgtype.Numeric{Int: big.NewInt(12340), Exp: -3, Valid: true}, `"12.34"`},
	}
	for _, test := range cases {
		value, err := Decode(fieldType, test.value)
		if err != nil {
			t.Fatal(err)
		}
		requireResponse(t, value, test.want)
	}
}

func TestDecodeDecimalRejectsOverflowPrecisionLossAndShapeDrift(t *testing.T) {
	fieldType := decimalType(t, 5, 2)
	cases := []any{
		"12.34", float64(12.34), int64(12),
		pgtype.Numeric{Int: big.NewInt(1234), Exp: -3, Valid: true},
		pgtype.Numeric{Int: big.NewInt(100000), Exp: -2, Valid: true},
		pgtype.Numeric{Int: big.NewInt(1), Exp: 1000, Valid: true},
		pgtype.Numeric{Int: big.NewInt(1), Exp: -1000, Valid: true},
		pgtype.Numeric{Int: big.NewInt(1), NaN: true, Valid: true},
		pgtype.Numeric{Int: big.NewInt(1), InfinityModifier: pgtype.NegativeInfinity, Valid: true},
		pgtype.Numeric{Valid: true},
	}
	for _, candidate := range cases {
		if _, err := Decode(fieldType, candidate); err == nil {
			t.Fatalf("invalid decimal %T accepted", candidate)
		}
	}
}
