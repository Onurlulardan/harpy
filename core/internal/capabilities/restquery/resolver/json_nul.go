package resolver

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
)

func jsonDocumentNULState(document []byte) (bool, bool) {
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return false, false
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return false, false
	}
	return jsonValueContainsNUL(value), true
}

func jsonValueContainsNUL(value any) bool {
	switch typed := value.(type) {
	case string:
		return strings.ContainsRune(typed, '\x00')
	case []any:
		for _, item := range typed {
			if jsonValueContainsNUL(item) {
				return true
			}
		}
	case map[string]any:
		for key, item := range typed {
			if strings.ContainsRune(key, '\x00') || jsonValueContainsNUL(item) {
				return true
			}
		}
	}
	return false
}
