package profile

import "strings"

func normalizeDecimalValue(field Field, document []byte) ([]byte, error) {
	value, err := decodeString(document)
	if err != nil {
		return nil, ErrValidationInvalid
	}
	negative, whole, fraction, valid := decimalParts(value)
	precision, scale := field.validation.decimalPrecision, field.validation.decimalScale
	wholeDigits := len(whole)
	if whole == "0" {
		wholeDigits = 0
	}
	if !valid || wholeDigits > precision-scale || len(fraction) > scale {
		return nil, ErrValidationInvalid
	}
	canonical := whole
	if scale > 0 {
		canonical += "." + fraction + strings.Repeat("0", scale-len(fraction))
	}
	if negative {
		canonical = "-" + canonical
	}
	return encodeScalar(canonical)
}

func decimalParts(value string) (bool, string, string, bool) {
	if value == "" || value[0] == '+' || strings.ContainsAny(value, "eE") {
		return false, "", "", false
	}
	negative := value[0] == '-'
	unsigned := strings.TrimPrefix(value, "-")
	parts := strings.Split(unsigned, ".")
	if len(parts) > 2 || parts[0] == "" || len(parts) == 2 && parts[1] == "" ||
		len(parts[0]) > 1 && parts[0][0] == '0' {
		return false, "", "", false
	}
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	for _, part := range parts {
		for _, digit := range part {
			if digit < '0' || digit > '9' {
				return false, "", "", false
			}
		}
	}
	if negative && !strings.ContainsAny(unsigned, "123456789") {
		return false, "", "", false
	}
	return negative, parts[0], fraction, true
}
