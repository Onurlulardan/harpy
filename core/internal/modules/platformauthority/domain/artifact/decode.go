package artifact

import (
	"bytes"
	"encoding/json"
	"io"
)

func strictCanonicalDecode(document []byte, target any) ([]byte, error) {
	canonical, err := canonicalDocument(document)
	if err != nil || !bytes.Equal(document, canonical) {
		return nil, ErrInvalidDocument
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return nil, ErrInvalidDocument
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, ErrInvalidDocument
	}
	return canonical, nil
}
