package artifact

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"sort"
	"unicode/utf8"
)

func canonicalDocument(document []byte) ([]byte, error) {
	if len(document) < 2 || len(document) > MaximumBytes || !utf8.Valid(document) {
		return nil, ErrInvalidDocument
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.UseNumber()
	value, err := scanCanonicalValue(decoder, 0)
	if err != nil {
		return nil, ErrInvalidDocument
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, ErrInvalidDocument
	}
	if _, object := value.(map[string]any); !object {
		return nil, ErrInvalidDocument
	}
	canonical, err := encodeCanonicalValue(nil, value)
	if err != nil || len(canonical) > MaximumBytes {
		return nil, ErrInvalidDocument
	}
	return canonical, nil
}

func canonicalValue(value any) ([]byte, error) {
	document, err := json.Marshal(value)
	if err != nil {
		return nil, ErrInvalidDocument
	}
	return canonicalDocument(document)
}

func digestHex(document []byte) string {
	digest := sha256.Sum256(document)
	return hex.EncodeToString(digest[:])
}

func signatureInput(domain string, canonicalUnsigned []byte) []byte {
	message := make([]byte, 0, len(domain)+1+len(canonicalUnsigned))
	message = append(message, domain...)
	message = append(message, 0)
	return append(message, canonicalUnsigned...)
}

func sortedKeys(value map[string]any) []string {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
