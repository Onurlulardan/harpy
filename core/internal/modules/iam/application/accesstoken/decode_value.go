package accesstoken

import (
	"encoding/json"
	"strconv"

	"github.com/google/uuid"
)

func decodeString(raw json.RawMessage) (string, bool) {
	var value string
	if len(raw) == 0 || string(raw) == "null" || json.Unmarshal(raw, &value) != nil {
		return "", false
	}
	return value, true
}

func decodePositiveInt64(raw json.RawMessage) (int64, bool) {
	if len(raw) == 0 || (len(raw) > 1 && raw[0] == '0') {
		return 0, false
	}
	for _, character := range raw {
		if character < '0' || character > '9' {
			return 0, false
		}
	}
	value, err := strconv.ParseInt(string(raw), 10, 64)
	return value, err == nil && value > 0
}

func decodeNonNegativeInt64(raw json.RawMessage) (int64, bool) {
	if len(raw) == 0 || (len(raw) > 1 && raw[0] == '0') {
		return 0, false
	}
	for _, character := range raw {
		if character < '0' || character > '9' {
			return 0, false
		}
	}
	value, err := strconv.ParseInt(string(raw), 10, 64)
	return value, err == nil && value >= 0
}

func decodeCanonicalUUID(raw json.RawMessage, requireV4 bool) (uuid.UUID, bool) {
	text, ok := decodeString(raw)
	if !ok {
		return uuid.Nil, false
	}
	value, err := uuid.Parse(text)
	if err != nil || value == uuid.Nil || value.String() != text {
		return uuid.Nil, false
	}
	if requireV4 && !validV4UUID(value) {
		return uuid.Nil, false
	}
	return value, true
}

func decodeMethods(raw json.RawMessage) ([]AuthMethod, bool) {
	var values []string
	if json.Unmarshal(raw, &values) != nil || values == nil {
		return nil, false
	}
	methods := make([]AuthMethod, len(values))
	for index, value := range values {
		methods[index] = AuthMethod(value)
	}
	return methods, validMethods(methods)
}
