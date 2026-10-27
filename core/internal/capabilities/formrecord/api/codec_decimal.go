package recordkernel

import (
	"encoding/json"
	"math/big"
	"regexp"
	"strings"
)

var decimalPattern = regexp.MustCompile(
	`^-?(0|[1-9][0-9]*)(\.[0-9]+)?$`,
)

func decodeDecimal(field Field, raw []byte) (Value, error) {
	text, err := decodeNumericText(raw)
	if field.Decimal == nil || err != nil ||
		!decimalPattern.MatchString(text) || negativeZero(text) {
		return Value{}, ErrInputInvalid
	}
	signless := strings.TrimPrefix(text, "-")
	parts := strings.SplitN(signless, ".", 2)
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	if len(fraction) > field.Decimal.Scale ||
		len(strings.TrimLeft(parts[0], "0")) > field.Decimal.Precision-field.Decimal.Scale {
		return Value{}, ErrInputInvalid
	}
	fraction += strings.Repeat("0", field.Decimal.Scale-len(fraction))
	canonical := parts[0]
	if field.Decimal.Scale > 0 {
		canonical += "." + fraction
	}
	if strings.HasPrefix(text, "-") {
		canonical = "-" + canonical
	}
	digits := strings.TrimLeft(parts[0], "0") + fraction
	if len(digits) > field.Decimal.Precision {
		return Value{}, ErrInputInvalid
	}
	encoded, _ := json.Marshal(canonical)
	result := typedValue(field.TypeID, encoded)
	result.decimal, _ = new(big.Rat).SetString(canonical)
	return result, nil
}

func negativeZero(value string) bool {
	if !strings.HasPrefix(value, "-") {
		return false
	}
	for _, character := range value[1:] {
		if character >= '1' && character <= '9' {
			return false
		}
	}
	return true
}
