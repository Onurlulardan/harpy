package profile

import (
	"bytes"
	"encoding/json"
	"strings"
)

func normalizeJSONValue(document []byte) ([]byte, error) {
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.UseNumber()
	var value any
	if decoder.Decode(&value) != nil || value == nil || !validJSONStrings(value) {
		return nil, ErrValidationInvalid
	}
	return document, nil
}

func validJSONStrings(value any) bool {
	switch typed := value.(type) {
	case string:
		return !strings.ContainsRune(typed, '\x00')
	case []any:
		for _, item := range typed {
			if !validJSONStrings(item) {
				return false
			}
		}
	case map[string]any:
		for key, item := range typed {
			if strings.ContainsRune(key, '\x00') || !validJSONStrings(item) {
				return false
			}
		}
	}
	return true
}
