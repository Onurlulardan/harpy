package config

import (
	"encoding/base64"
	"encoding/json"
	"unicode/utf8"
)

type cursorKeyDocument struct {
	kid string
	key string
}

func loadCursorKeyring(values environmentValues) (CursorKeyring, error) {
	secret, err := loadSecret(values, cursorKeyringEnvironmentKey, true, secretFileLimit)
	if err != nil {
		return CursorKeyring{}, err
	}
	return parseCursorKeyring(secret)
}

func parseCursorKeyring(secret Secret) (CursorKeyring, error) {
	document := []byte(secret.Value())
	if !utf8.Valid(document) || rejectDuplicateJSONMembers(document) != nil {
		return CursorKeyring{}, invalidCursorKeyring()
	}
	var root map[string]json.RawMessage
	if json.Unmarshal(document, &root) != nil || !hasExactMembers(root, "version", "current", "previous") {
		return CursorKeyring{}, invalidCursorKeyring()
	}
	var version int
	if json.Unmarshal(root["version"], &version) != nil || version != 1 {
		return CursorKeyring{}, invalidCursorKeyring()
	}
	currentDocument, ok := decodeCursorKeyDocument(root["current"])
	if !ok {
		return CursorKeyring{}, invalidCursorKeyring()
	}
	previousDocument, ok := decodeCursorKeyDocument(root["previous"])
	if !ok || currentDocument.kid == previousDocument.kid {
		return CursorKeyring{}, invalidCursorKeyring()
	}
	current, ok := decodeCursorKey(currentDocument)
	if !ok {
		return CursorKeyring{}, invalidCursorKeyring()
	}
	previous, ok := decodeCursorKey(previousDocument)
	if !ok {
		return CursorKeyring{}, invalidCursorKeyring()
	}
	return CursorKeyring{current: current, previous: previous, valid: true}, nil
}

func decodeCursorKeyDocument(raw json.RawMessage) (cursorKeyDocument, bool) {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil || !hasExactMembers(object, "kid", "key") {
		return cursorKeyDocument{}, false
	}
	var document cursorKeyDocument
	if json.Unmarshal(object["kid"], &document.kid) != nil ||
		json.Unmarshal(object["key"], &document.key) != nil {
		return cursorKeyDocument{}, false
	}
	return document, true
}

func hasExactMembers(object map[string]json.RawMessage, names ...string) bool {
	if len(object) != len(names) {
		return false
	}
	for _, name := range names {
		if _, exists := object[name]; !exists {
			return false
		}
	}
	return true
}

func decodeCursorKey(document cursorKeyDocument) (CursorKey, bool) {
	if !validCursorKID(document.kid) {
		return CursorKey{}, false
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(document.key)
	if err != nil || len(decoded) != cursorKeySize ||
		base64.RawURLEncoding.EncodeToString(decoded) != document.key {
		clear(decoded)
		return CursorKey{}, false
	}
	var material [cursorKeySize]byte
	copy(material[:], decoded)
	clear(decoded)
	return CursorKey{kid: document.kid, material: material, valid: true}, true
}

func validCursorKID(value string) bool {
	if len(value) < 1 || len(value) > 32 || !lowercaseLetterOrDigit(value[0]) {
		return false
	}
	for index := 1; index < len(value); index++ {
		if !lowercaseLetterOrDigit(value[index]) && value[index] != '_' && value[index] != '-' {
			return false
		}
	}
	return true
}

func lowercaseLetterOrDigit(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= '0' && value <= '9'
}

func invalidCursorKeyring() error {
	return configError(cursorKeyringEnvironmentKey, "must satisfy rest-query-cursor-keyring.v1")
}
