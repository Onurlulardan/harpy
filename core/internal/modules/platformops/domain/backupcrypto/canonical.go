package backupcrypto

import (
	"bytes"
	"encoding/json"
)

func canonicalObject(value any, maximum int) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, ErrInvalidArchive
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var document map[string]any
	if decoder.Decode(&document) != nil || document == nil {
		return nil, ErrInvalidArchive
	}
	canonical, err := json.Marshal(document)
	if err != nil || len(canonical) > maximum {
		return nil, ErrInvalidArchive
	}
	return canonical, nil
}
