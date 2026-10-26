package recoveryset

import (
	"bytes"
	"encoding/json"
	"io"
)

var componentNames = map[string]bool{
	"database_archive":    true,
	"documents_inventory": true,
	"key_custody":         true,
	"owner_anchor":        true,
	"protected_authority": true,
}

func ParseComponentInventory(
	document []byte,
	requireComplete bool,
) (map[string]ComponentIdentity, error) {
	if len(document) < 2 || len(document) > MaximumManifestBytes {
		return nil, ErrManifest
	}
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(document))
	if decoder.Decode(&raw) != nil || decoder.Decode(new(any)) != io.EOF ||
		len(raw) > len(componentNames) ||
		requireComplete && len(raw) != len(componentNames) {
		return nil, ErrManifest
	}
	result := make(map[string]ComponentIdentity, len(raw))
	for name, encoded := range raw {
		if !componentNames[name] {
			return nil, ErrManifest
		}
		var component ComponentIdentity
		itemDecoder := json.NewDecoder(bytes.NewReader(encoded))
		itemDecoder.DisallowUnknownFields()
		if itemDecoder.Decode(&component) != nil ||
			itemDecoder.Decode(new(any)) != io.EOF || !component.valid() {
			return nil, ErrManifest
		}
		result[name] = component
	}
	return result, nil
}
