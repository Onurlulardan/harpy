package attestation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

// CanonicalizeSchemaPlan applies the schema-plan contract's exact byte budget.
func CanonicalizeSchemaPlan(document []byte) ([]byte, error) {
	return canonicalizeBounded(document, schemav1.MaximumSchemaPlanBytes)
}

// MarshalCanonicalSchemaPlan encodes a typed plan value as RFC 8785 JSON.
func MarshalCanonicalSchemaPlan(value any) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode schema plan JSON: %w", err)
	}
	return CanonicalizeSchemaPlan(encoded)
}

// DigestSchemaPlan hashes the canonical bytes of a schema-plan document.
func DigestSchemaPlan(document []byte) (string, error) {
	canonical, err := CanonicalizeSchemaPlan(document)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:]), nil
}
