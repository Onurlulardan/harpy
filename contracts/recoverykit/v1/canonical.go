package recoverykitv1

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

const canonicalTimeFormat = time.RFC3339Nano

func canonicalObject(value any) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("recovery kit encoding failed")
	}
	defer clear(raw)
	return canonicaljsonv1.CanonicalizeObject(raw, MaximumKitBytes)
}

func parseCanonical[T any](document []byte, target *T) error {
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumKitBytes)
	if err != nil || !bytes.Equal(canonical, document) {
		return fmt.Errorf("recovery kit document is not canonical")
	}
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.DisallowUnknownFields()
	if decoder.Decode(target) != nil {
		return fmt.Errorf("recovery kit document is invalid")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return fmt.Errorf("recovery kit document has trailing data")
	}
	return nil
}

func encode(value []byte) string { return base64.RawURLEncoding.EncodeToString(value) }

func decode(value string, size int) ([]byte, bool) {
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	valid := err == nil && (size < 0 || len(decoded) == size) &&
		base64.RawURLEncoding.EncodeToString(decoded) == value
	return decoded, valid
}

func digestHex(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}

func canonicalTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1000 == 0
}
