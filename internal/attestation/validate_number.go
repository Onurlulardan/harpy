package attestation

import (
	"math"
	"math/big"
	"strconv"
	"strings"
)

var maximumSafeJSONInteger = big.NewInt(9007199254740991)

func validateJSONNumber(number string) error {
	if strings.ContainsAny(number, ".eE") {
		value, err := strconv.ParseFloat(number, 64)
		if err != nil || math.IsInf(value, 0) || math.IsNaN(value) {
			return errInvalidJSONNumber
		}
		return nil
	}
	integer, valid := new(big.Int).SetString(number, 10)
	if !valid {
		return errInvalidJSONNumber
	}
	absolute := new(big.Int).Abs(integer)
	if absolute.Cmp(maximumSafeJSONInteger) > 0 {
		return errInvalidJSONNumber
	}
	return nil
}
