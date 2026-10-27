package recordkernel

import (
	"encoding/json"
	"github.com/Onurlulardan/harpy/core/internal/platform/textprofile"
	"math/big"
	"strconv"
)

func buildText(value string) (Value, error) {
	encoded, err := json.Marshal(textprofile.ExactNFC(value))
	if err != nil || len(encoded) > MaximumFieldBytes {
		return Value{}, ErrBudgetExceeded
	}
	return typedValue(TypeText, encoded), nil
}

func buildInteger(value int64) Value {
	encoded, _ := json.Marshal(strconv.FormatInt(value, 10))
	return typedValue(TypeInteger, encoded)
}

func buildBoolean(value bool) Value {
	if value {
		return typedValue(TypeBoolean, []byte("true"))
	}
	return typedValue(TypeBoolean, []byte("false"))
}

func buildRoundedDecimal(value *big.Rat, profile DecimalProfile) (Value, error) {
	text, rounded, err := roundedDecimal(value, profile.Scale, profile.Precision)
	if err != nil {
		return Value{}, err
	}
	encoded, _ := json.Marshal(text)
	result := typedValue(TypeDecimal, encoded)
	result.decimal = rounded
	return result, nil
}
