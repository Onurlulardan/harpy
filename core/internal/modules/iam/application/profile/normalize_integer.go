package profile

import (
	"strconv"
	"strings"
)

func normalizeIntegerValue(field Field, document []byte) ([]byte, error) {
	value, err := decodeString(document)
	if err != nil || !canonicalInteger(value) {
		return nil, ErrValidationInvalid
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < field.validation.integerMinimum ||
		parsed > field.validation.integerMaximum {
		return nil, ErrValidationInvalid
	}
	return encodeScalar(value)
}

func canonicalInteger(value string) bool {
	if value == "" || value == "-0" || value[0] == '+' ||
		len(value) > 1 && value[0] == '0' ||
		len(value) > 1 && value[0] == '-' && value[1] == '0' {
		return false
	}
	digits := strings.TrimPrefix(value, "-")
	if digits == "" {
		return false
	}
	for _, digit := range digits {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}
