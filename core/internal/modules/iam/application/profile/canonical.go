package profile

import (
	"crypto/sha256"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func canonicalMarshal(value any, maximumBytes int) ([]byte, error) {
	document, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return canonicaljsonv1.Canonicalize(document, maximumBytes)
}

func canonicalObject(document []byte) ([]byte, error) {
	return canonicaljsonv1.CanonicalizeObject(document, MaximumDocumentBytes)
}

func canonicalDigest(document []byte) [sha256.Size]byte {
	return sha256.Sum256(document)
}
