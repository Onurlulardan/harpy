package recoveryset

import (
	"bytes"
	"encoding/json"
)

func canonicalObject(value any, maximum int) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, ErrManifest
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var document map[string]any
	if decoder.Decode(&document) != nil || document == nil {
		return nil, ErrManifest
	}
	canonical, err := json.Marshal(document)
	if err != nil || len(canonical) > maximum {
		return nil, ErrManifest
	}
	return canonical, nil
}
