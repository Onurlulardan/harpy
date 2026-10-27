package resolved

import (
	"bytes"
	"encoding/json"
	"strings"
)

func validJSONStrings(document []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.UseNumber()
	var value any
	if decoder.Decode(&value) != nil {
		return false
	}
	return jsonValueStringsValid(value)
}

func jsonValueStringsValid(value any) bool {
	switch typed := value.(type) {
	case string:
		return !strings.ContainsRune(typed, '\x00')
	case []any:
		for _, item := range typed {
			if !jsonValueStringsValid(item) {
				return false
			}
		}
	case map[string]any:
		for key, item := range typed {
			if strings.ContainsRune(key, '\x00') || !jsonValueStringsValid(item) {
				return false
			}
		}
	}
	return true
}
