package recordkernel

import (
	"encoding/json"
	"regexp"
	"strconv"
)

var integerPattern = regexp.MustCompile(`^(0|-[1-9][0-9]*|[1-9][0-9]*)$`)

func decodeInteger(typeID TypeID, raw []byte) (Value, error) {
	text, err := decodeNumericText(raw)
	if err != nil || !integerPattern.MatchString(text) {
		return Value{}, ErrInputInvalid
	}
	if _, err := strconv.ParseInt(text, 10, 64); err != nil {
		return Value{}, ErrInputInvalid
	}
	encoded, _ := json.Marshal(text)
	return typedValue(typeID, encoded), nil
}

func decodeNumericText(raw []byte) (string, error) {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text, nil
	}
	var number json.Number
	if json.Unmarshal(raw, &number) != nil || string(number) == "" {
		return "", ErrInputInvalid
	}
	return string(number), nil
}
