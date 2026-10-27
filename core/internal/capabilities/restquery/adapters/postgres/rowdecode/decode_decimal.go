package rowdecode

import (
	"math/big"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/jackc/pgx/v5/pgtype"
)

func decodeDecimal(fieldType api.FieldType, driverValue any) (Value, error) {
	numeric, ok := driverValue.(pgtype.Numeric)
	if !ok || !finiteNumeric(numeric) {
		return Value{}, ErrInvalidValue
	}
	profile, found, err := fieldType.Decimal()
	if err != nil || !found {
		return Value{}, ErrInvalidFieldType
	}
	precision, precisionErr := profile.Precision()
	scale, scaleErr := profile.Scale()
	if precisionErr != nil || scaleErr != nil {
		return Value{}, ErrInvalidFieldType
	}
	canonical, ok := decimalString(numeric, precision, scale)
	if !ok {
		return Value{}, ErrInvalidValue
	}
	return scalarValue(fieldType, api.LogicalTypeDecimal, canonical)
}

func finiteNumeric(value pgtype.Numeric) bool {
	return value.Valid && value.Int != nil && !value.NaN &&
		value.InfinityModifier == pgtype.Finite
}

func decimalString(value pgtype.Numeric, precision, scale int) (string, bool) {
	scaled := new(big.Int).Set(value.Int)
	shift := int64(value.Exp) + int64(scale)
	if !applyDecimalShift(scaled, shift, precision) {
		return "", false
	}
	digits := new(big.Int).Abs(new(big.Int).Set(scaled)).String()
	if len(digits) > precision {
		return "", false
	}
	if scale > 0 && len(digits) <= scale {
		digits = strings.Repeat("0", scale-len(digits)+1) + digits
	}
	if scale > 0 {
		split := len(digits) - scale
		digits = digits[:split] + "." + digits[split:]
	}
	if scaled.Sign() < 0 {
		digits = "-" + digits
	}
	return digits, true
}

func applyDecimalShift(value *big.Int, shift int64, precision int) bool {
	if value.Sign() == 0 {
		value.SetInt64(0)
		return true
	}
	if shift >= 0 {
		if shift > int64(precision) || len(new(big.Int).Abs(value).String())+int(shift) > precision {
			return false
		}
		value.Mul(value, powerOfTen(int(shift)))
		return true
	}
	exponent := -shift
	if exponent > int64(precision)+1 {
		return false
	}
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(value, powerOfTen(int(exponent)), remainder)
	if remainder.Sign() != 0 {
		return false
	}
	value.Set(quotient)
	return true
}

func powerOfTen(exponent int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(exponent)), nil)
}
