package rowdecode

import (
	"math/big"
	"strconv"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/jackc/pgx/v5/pgtype"
)

func decodeInteger(fieldType api.FieldType, driverValue any) (Value, error) {
	integer, ok := driverInteger(driverValue)
	if !ok {
		return Value{}, ErrInvalidValue
	}
	profile, found, err := fieldType.Integer()
	if err != nil || !found {
		return Value{}, ErrInvalidFieldType
	}
	minimum, minimumErr := profile.Minimum()
	maximum, maximumErr := profile.Maximum()
	if minimumErr != nil || maximumErr != nil || integer < minimum || integer > maximum {
		return Value{}, ErrInvalidValue
	}
	return scalarValue(
		fieldType, api.LogicalTypeInteger, strconv.FormatInt(integer, 10),
	)
}

func driverInteger(driverValue any) (int64, bool) {
	switch value := driverValue.(type) {
	case int16:
		return int64(value), true
	case int32:
		return int64(value), true
	case int64:
		return value, true
	case pgtype.Numeric:
		return integralNumeric(value)
	default:
		return 0, false
	}
}

func integralNumeric(value pgtype.Numeric) (int64, bool) {
	if !finiteNumeric(value) {
		return 0, false
	}
	integer := new(big.Int).Set(value.Int)
	if value.Exp > 0 {
		if value.Exp > 18 {
			return 0, false
		}
		integer.Mul(integer, powerOfTen(int(value.Exp)))
	} else if value.Exp < 0 {
		exponent := -int64(value.Exp)
		if exponent > 19 {
			return 0, false
		}
		divisor := powerOfTen(int(exponent))
		quotient, remainder := new(big.Int), new(big.Int)
		quotient.QuoRem(integer, divisor, remainder)
		if remainder.Sign() != 0 {
			return 0, false
		}
		integer = quotient
	}
	return integer.Int64(), integer.IsInt64()
}
