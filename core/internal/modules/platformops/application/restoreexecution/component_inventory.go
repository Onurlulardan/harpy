package restoreexecution

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

var requiredComponents = map[string]struct{}{
	"database_archive":    {},
	"documents_inventory": {},
	"key_custody":         {},
	"owner_anchor":        {},
	"protected_authority": {},
}

type componentInventoryItem struct {
	ObjectKey        string `json:"object_key"`
	ObjectVersionID  string `json:"object_version_id"`
	CiphertextSHA256 string `json:"ciphertext_sha256"`
	CiphertextBytes  int64  `json:"ciphertext_bytes"`
}

func validComponentInventory(document json.RawMessage) bool {
	var components map[string]json.RawMessage
	if json.Unmarshal(document, &components) != nil ||
		len(components) != len(requiredComponents) {
		return false
	}
	for name := range requiredComponents {
		raw, found := components[name]
		if !found || !validComponent(raw) {
			return false
		}
	}
	return true
}

func validComponent(document []byte) bool {
	var item componentInventoryItem
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&item) != nil || decoder.Decode(new(any)) != io.EOF {
		return false
	}
	return len(item.ObjectKey) > 0 && len(item.ObjectKey) <= 1024 &&
		!strings.ContainsAny(item.ObjectKey, "\x00\r\n") &&
		len(item.ObjectVersionID) > 0 && len(item.ObjectVersionID) <= 1024 &&
		!strings.ContainsAny(item.ObjectVersionID, "\x00\r\n") &&
		digestPattern.MatchString(item.CiphertextSHA256) &&
		item.CiphertextBytes > 0
}
