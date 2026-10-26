package canonicaljsonv1

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"unicode/utf8"

	"github.com/gowebpki/jcs"
)

// Canonicalize validates and serializes one JSON value according to RFC 8785.
func Canonicalize(document []byte, maximumBytes int) ([]byte, error) {
	if maximumBytes < 1 || len(document) < 1 || len(document) > maximumBytes ||
		!utf8.Valid(document) {
		return nil, fmt.Errorf("canonical JSON size or UTF-8 is invalid")
	}
	if err := Validate(document); err != nil {
		return nil, fmt.Errorf("canonical JSON input is ambiguous: %w", err)
	}
	canonical, err := jcs.Transform(document)
	if err != nil {
		return nil, fmt.Errorf("canonicalize RFC 8785 JSON: %w", err)
	}
	if len(canonical) > maximumBytes {
		return nil, fmt.Errorf("canonical JSON output exceeds its byte limit")
	}
	return canonical, nil
}

// CanonicalizeObject requires the canonical root value to be an object.
func CanonicalizeObject(document []byte, maximumBytes int) ([]byte, error) {
	canonical, err := Canonicalize(document, maximumBytes)
	if err != nil {
		return nil, err
	}
	if len(canonical) < 2 || canonical[0] != '{' {
		return nil, fmt.Errorf("canonical JSON root must be an object")
	}
	return canonical, nil
}

// DigestHex returns SHA-256 over the RFC 8785 canonical bytes.
func DigestHex(document []byte, maximumBytes int) (string, error) {
	canonical, err := Canonicalize(document, maximumBytes)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:]), nil
}
