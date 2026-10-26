package recordkernel

import (
	"encoding/json"
	"github.com/Onurlulardan/harpy/core/internal/platform/textprofile"
	"strings"
)

func decodeText(typeID TypeID, raw []byte) (Value, error) {
	var text string
	if json.Unmarshal(raw, &text) != nil || strings.ContainsRune(text, '\x00') {
		return Value{}, ErrInputInvalid
	}
	encoded, err := json.Marshal(textprofile.ExactNFC(text))
	if err != nil {
		return Value{}, ErrInputInvalid
	}
	return typedValue(typeID, encoded), nil
}

func decodeBoolean(typeID TypeID, raw []byte) (Value, error) {
	if string(raw) != "true" && string(raw) != "false" {
		return Value{}, ErrInputInvalid
	}
	return typedValue(typeID, raw), nil
}
