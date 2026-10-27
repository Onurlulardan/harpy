package attestation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"unicode/utf8"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/gowebpki/jcs"
)

const maximumDocumentBytes = runtimev1.MaximumAttestationDocumentBytes

func Canonicalize(document []byte) ([]byte, error) {
	return canonicalizeBounded(document, maximumDocumentBytes)
}

func canonicalizeBounded(document []byte, maximumBytes int) ([]byte, error) {
	if !withinDocumentSizeLimit(len(document), maximumBytes) || !utf8.Valid(document) {
		return nil, fmt.Errorf("attestation document size is invalid")
	}
	if err := rejectDuplicateKeys(document); err != nil {
		return nil, fmt.Errorf("attestation JSON is not strict: %w", err)
	}
	canonical, err := jcs.Transform(document)
	if err != nil {
		return nil, fmt.Errorf("canonicalize attestation JSON: %w", err)
	}
	if len(canonical) > maximumBytes {
		return nil, fmt.Errorf("canonical attestation document size is invalid")
	}
	return canonical, nil
}

func MarshalCanonical(value any) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode attestation JSON: %w", err)
	}
	return Canonicalize(encoded)
}

func DigestHex(document []byte) (string, error) {
	canonical, err := Canonicalize(document)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:]), nil
}
